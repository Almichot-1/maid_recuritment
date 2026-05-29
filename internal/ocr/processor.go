package ocr

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// VisualZoneData holds fields extracted from the visual zone of a passport
// (the human-readable area above the MRZ).
type VisualZoneData struct {
	PlaceOfBirth     string    `json:"place_of_birth"`
	PlaceOfBirthConf float64   `json:"place_of_birth_conf"`
	DateOfIssue      time.Time `json:"date_of_issue"`
	DateOfIssueConf  float64   `json:"date_of_issue_conf"`
	Authority        string    `json:"authority"`
	AuthorityConf    float64   `json:"authority_conf"`
	RawText          string    `json:"raw_text"`
}

// OCRProcessor orchestrates Tesseract-based passport OCR.
// It extracts MRZ lines and the visual zone from a passport image.
type OCRProcessor struct {
	lang      string
	tessdata  string
	parser    *MRZParser
	tempFiles []string
}

// NewOCRProcessor creates a new OCRProcessor.
// lang is the Tesseract language code (e.g. "eng").
// tessdataPath is the TESSDATA_PREFIX directory; pass "" to use the system default.
func NewOCRProcessor(lang, tessdataPath string) *OCRProcessor {
	if strings.TrimSpace(lang) == "" {
		lang = "eng"
	}
	return &OCRProcessor{
		lang:     lang,
		tessdata: strings.TrimSpace(tessdataPath),
		parser:   NewMRZParser(),
	}
}

// ExtractMRZ runs Tesseract on the MRZ region of the image and returns the two
// MRZ lines together with a best-effort confidence value (0 when unavailable).
func (p *OCRProcessor) ExtractMRZ(imagePath string) (line1, line2 string, confidence float64, err error) {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return "", "", 0, fmt.Errorf("image path is required")
	}
	if _, statErr := os.Stat(imagePath); statErr != nil {
		return "", "", 0, fmt.Errorf("image file not found: %w", statErr)
	}

	text, runErr := p.runTesseract(imagePath, "ocrb", []string{
		"--psm", "6",
		"-c", "tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<",
	})
	if runErr != nil {
		return "", "", 0, runErr
	}

	l1, l2 := extractMRZLinesFromText(text)
	if l1 == "" || l2 == "" {
		return "", "", 0, fmt.Errorf("MRZ OCR produced invalid output: expected 2 MRZ lines")
	}
	return l1, l2, 0, nil
}

// ExtractVisualZone runs Tesseract on the full image with two PSM passes to
// extract Place of Birth, Issue Date, and Authority.
func (p *OCRProcessor) ExtractVisualZone(imagePath string) (*VisualZoneData, error) {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return nil, fmt.Errorf("image path is required")
	}
	if _, statErr := os.Stat(imagePath); statErr != nil {
		return nil, fmt.Errorf("image file not found: %w", statErr)
	}

	common := []string{
		"--oem", "1",
		"-c", "preserve_interword_spaces=1",
		"-c", "user_defined_dpi=300",
	}

	// Two-pass OCR: different PSMs often recover different lines on passports.
	textPSM6, err6 := p.runTesseract(imagePath, p.lang, append([]string{"--psm", "6"}, common...))
	textPSM11, err11 := p.runTesseract(imagePath, p.lang, append([]string{"--psm", "11"}, common...))
	if err6 != nil && err11 != nil {
		return nil, err6
	}
	text := mergeOCRTextLines(textPSM6, textPSM11)

	vz := extractVisualZoneFields(text, 0)
	vz.RawText = text
	return vz, nil
}

// SetLanguage switches the OCR language for subsequent calls.
// Returns an error when the language is not available in the Tesseract install.
func (p *OCRProcessor) SetLanguage(lang string) error {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return fmt.Errorf("language is required")
	}

	langs, err := p.listLangs()
	if err != nil {
		// If we cannot query languages we proceed optimistically.
		p.lang = lang
		return nil
	}
	if len(langs) > 0 {
		found := false
		for _, l := range langs {
			if strings.EqualFold(l, lang) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("tesseract language %q is not installed", lang)
		}
	}

	p.lang = lang
	return nil
}

// Close cleans up any temporary files created during processing.
func (p *OCRProcessor) Close() error {
	for _, f := range p.tempFiles {
		_ = os.Remove(f)
	}
	p.tempFiles = nil
	return nil
}

// ─── internal ────────────────────────────────────────────────────────────────

func (p *OCRProcessor) trackTempFile(path string) {
	if path = strings.TrimSpace(path); path != "" {
		p.tempFiles = append(p.tempFiles, path)
	}
}

// runTesseract invokes the Tesseract CLI and returns stdout as a string.
func (p *OCRProcessor) runTesseract(imagePath, lang string, extraArgs []string) (string, error) {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return "", fmt.Errorf("image path is required")
	}

	args := []string{imagePath, "stdout"}
	args = append(args, extraArgs...)
	args = append(args, "-l", lang)

	cmd := exec.Command(p.tesseractExe(), args...)
	cmd.Env = append([]string{}, os.Environ()...)
	cmd.Env = p.appendTessdataEnv(cmd.Env)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("tesseract failed: %s", msg)
	}

	return stdout.String(), nil
}

// listLangs queries the installed Tesseract languages.
func (p *OCRProcessor) listLangs() ([]string, error) {
	cmd := exec.Command(p.tesseractExe(), "--list-langs")
	cmd.Env = append([]string{}, os.Environ()...)
	cmd.Env = p.appendTessdataEnv(cmd.Env)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("list tesseract languages: %w", err)
	}

	lines := strings.Split(strings.ReplaceAll(string(out), "\r", ""), "\n")
	langs := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(l), "list of available") {
			continue
		}
		langs = append(langs, l)
	}
	return langs, nil
}

// tesseractExe resolves the Tesseract executable path.
// It prefers PATH lookup, then falls back to standard Windows install locations.
func (p *OCRProcessor) tesseractExe() string {
	if path, err := exec.LookPath("tesseract"); err == nil && strings.TrimSpace(path) != "" {
		return path
	}
	if runtime.GOOS == "windows" {
		candidates := []string{
			`C:\Program Files\Tesseract-OCR\tesseract.exe`,
			`C:\Program Files (x86)\Tesseract-OCR\tesseract.exe`,
		}
		for _, c := range candidates {
			if fileExistsOCR(c) {
				return c
			}
		}
	}
	return "tesseract"
}

// appendTessdataEnv adds TESSDATA_PREFIX to the environment when a valid
// tessdata directory is configured.
func (p *OCRProcessor) appendTessdataEnv(env []string) []string {
	td := p.tessdata
	if td == "" {
		return env
	}
	if info, err := os.Stat(td); err == nil && info != nil && info.IsDir() {
		return append(env, "TESSDATA_PREFIX="+td)
	}
	if info, err := os.Stat(filepath.Join(td, "tessdata")); err == nil && info != nil && info.IsDir() {
		return append(env, "TESSDATA_PREFIX="+td)
	}
	return env
}

// mergeOCRTextLines merges multiple OCR text outputs, deduplicating lines
// (case-insensitive) while preserving order.
func mergeOCRTextLines(texts ...string) string {
	seen := make(map[string]struct{}, 256)
	out := make([]string, 0, 128)
	for _, t := range texts {
		t = strings.ReplaceAll(t, "\r", "")
		for _, raw := range strings.Split(t, "\n") {
			l := strings.TrimSpace(raw)
			if l == "" {
				continue
			}
			k := strings.ToUpper(l)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

func fileExistsOCR(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info != nil && !info.IsDir()
}
