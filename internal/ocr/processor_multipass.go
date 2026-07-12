package ocr

import (
	"sync"
	"time"
)

const maxConcurrentMRZPasses = 3

// MRZPassConfig defines a single Tesseract pass configuration.
type MRZPassConfig struct {
	PSMMode int
	OEMMode int
	Label   string
}

// MRZPassResult holds the outcome of a single Tesseract pass.
type MRZPassResult struct {
	Config     MRZPassConfig
	Language   string
	RawText    string
	Line1      string
	Line2      string
	Parsed     *MRZData
	Confidence float64
	Err        error
}

// defaultMRZPassConfigs returns the default set of MRZ pass configurations.
func defaultMRZPassConfigs() []MRZPassConfig {
	return []MRZPassConfig{
		{PSMMode: 6, OEMMode: 1, Label: "psm6_oem1"},
		{PSMMode: 3, OEMMode: 1, Label: "psm3_oem1"},
		{PSMMode: 11, OEMMode: 1, Label: "psm11_oem1"},
		{PSMMode: 6, OEMMode: 0, Label: "psm6_oem0"},
		{PSMMode: 6, OEMMode: 2, Label: "psm6_oem2"},
	}
}

// tryOCROnImage runs Tesseract on a single image path and populates the result.
func tryOCROnImage(p *OCRProcessor, imagePath, language string, config MRZPassConfig, args []string, timeout time.Duration, result *MRZPassResult) bool {
	text, err := p.runTesseractTextWithTimeout(imagePath, language, args, timeout)
	if err != nil {
		return false
	}

	line1, line2 := extractMRZLinesFromText(text)
	if line1 == "" || line2 == "" {
		return false
	}

	parsed, parseErr := p.parser.ParseMRZ(line1, line2)
	if parseErr != nil {
		return false
	}

	conf := ComputeConfidence(parsed)
	// Only keep if better than what we already have
	if result.Parsed == nil || conf > result.Confidence {
		result.RawText = text
		result.Line1 = line1
		result.Line2 = line2
		result.Parsed = parsed
		result.Confidence = conf
		result.Err = nil
	}
	return true
}

// runSingleMRZPass executes one Tesseract pass with the given config and language.
// It tries both the raw image and a preprocessed version, keeping whichever yields
// better OCR results.
func runSingleMRZPass(p *OCRProcessor, imagePath, language string, config MRZPassConfig, timeout time.Duration) *MRZPassResult {
	result := &MRZPassResult{
		Config:   config,
		Language: language,
	}

	args := []string{
		"--oem", formatOEM(config.OEMMode),
		"--psm", formatPSM(config.PSMMode),
		"-c", "load_system_dawg=0",
		"-c", "load_freq_dawg=0",
		"-c", "user_defined_dpi=300",
		"-c", "tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<",
	}

	// Try raw image first (fast path, no preprocessing overhead)
	rawOK := tryOCROnImage(p, imagePath, language, config, args, timeout, result)

	// Try preprocessed image (may help on noisy/blurry images)
	procPath, cleanup, procErr := prepareMRZImage(imagePath)
	if procErr == nil {
		defer cleanup()
		procOK := tryOCROnImage(p, procPath, language, config, args, timeout, result)
		if !rawOK && !procOK {
			result.Err = ErrInvalidMRZOutput
		}
	} else if !rawOK {
		result.Err = ErrInvalidMRZOutput
	}

	return result
}

// runAllMRZPasses runs all pass configs across all languages concurrently.
func runAllMRZPasses(p *OCRProcessor, imagePath string, languages []string, configs []MRZPassConfig, timeout time.Duration) []*MRZPassResult {
	total := len(languages) * len(configs)
	results := make([]*MRZPassResult, total)
	var wg sync.WaitGroup
	wg.Add(total)

	sem := make(chan struct{}, maxConcurrentMRZPasses)
	idx := 0
	for _, lang := range languages {
		for _, cfg := range configs {
			i := idx
			sem <- struct{}{}
			go func(language string, config MRZPassConfig) {
				defer wg.Done()
				defer func() { <-sem }()
				results[i] = runSingleMRZPass(p, imagePath, language, config, timeout)
			}(lang, cfg)
			idx++
		}
	}

	wg.Wait()
	return results
}

// bestSinglePassResult returns the pass result with the highest confidence
// among successful passes (no error, parsed != nil). Returns nil if none.
func bestSinglePassResult(results []*MRZPassResult) *MRZPassResult {
	var best *MRZPassResult
	for _, r := range results {
		if r.Err != nil || r.Parsed == nil {
			continue
		}
		if best == nil || r.Confidence > best.Confidence {
			best = r
		}
	}
	return best
}

func formatOEM(mode int) string {
	switch mode {
	case 0:
		return "0"
	case 1:
		return "1"
	case 2:
		return "2"
	default:
		return "1"
	}
}

func formatPSM(mode int) string {
	switch mode {
	case 3:
		return "3"
	case 4:
		return "4"
	case 6:
		return "6"
	case 7:
		return "7"
	case 11:
		return "11"
	case 13:
		return "13"
	default:
		return "6"
	}
}
