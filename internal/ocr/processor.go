package ocr

import (
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

type OCRProcessor struct {
	tesseractPath string
	lang          string
	parser        *MRZParser
}

type VisualZoneData struct {
	PlaceOfBirth     string    `json:"place_of_birth"`
	PlaceOfBirthConf float64   `json:"place_of_birth_conf"`
	DateOfIssue      time.Time `json:"date_of_issue"`
	DateOfIssueConf  float64   `json:"date_of_issue_conf"`
	Authority        string    `json:"authority"`
	AuthorityConf    float64   `json:"authority_conf"`
	RawText          string    `json:"raw_text"`
}

type PassportData struct {
	DocumentType     string
	CountryCode      string
	Surname          string
	GivenNames       string
	PassportNumber   string
	Nationality      string
	DateOfBirth      time.Time
	Sex              string
	PlaceOfBirth     string
	DateOfIssue      time.Time
	DateOfExpiry     time.Time
	IssuingAuthority string
	MRZLine1         string
	MRZLine2         string
	Confidence       float64
	ExtractedAt      time.Time
}

type visualZoneExtractionResult struct {
	data *VisualZoneData
	err  error
}

func NewOCRProcessor(tesseractPath, lang string) *OCRProcessor {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		lang = "eng"
	}
	return &OCRProcessor{
		tesseractPath: strings.TrimSpace(tesseractPath),
		lang:          lang,
		parser:        NewMRZParser(),
	}
}

func (p *OCRProcessor) ExtractMRZ(imagePath string) (string, string, float64, error) {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return "", "", 0, os.ErrInvalid
	}
	if _, err := os.Stat(imagePath); err != nil {
		return "", "", 0, err
	}

	// Each pass handles its own preprocessing internally;
	// we pass the original raw path so each pass can try both
	// raw and preprocessed and keep the better result.
	languages := uniqueOCRLanguages("ocrb", "eng", p.lang)
	configs := defaultMRZPassConfigs()
	timeout := 20 * time.Second

	results := runAllMRZPasses(p, imagePath, languages, configs, timeout)

	for _, r := range results {
		if r.Err == nil && r.Parsed != nil && r.Parsed.IsValid && r.Confidence >= 0.5 {
			return r.Line1, r.Line2, r.Confidence, nil
		}
	}

	fused := fuseMRZResults(results)

	fused = rerunLowConfidenceFields(p, imagePath, fused)

	if fused != nil && fused.SuccessfulPasses > 0 && fused.Confidence > 0.3 {
		return fused.Line1, fused.Line2, fused.Confidence, nil
	}

	if best := bestSinglePassResult(results); best != nil {
		return best.Line1, best.Line2, best.Confidence, nil
	}

	return "", "", 0, ErrInvalidMRZOutput
}

func (p *OCRProcessor) ExtractVisualZone(imagePath string) (*VisualZoneData, error) {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return nil, os.ErrInvalid
	}
	if _, err := os.Stat(imagePath); err != nil {
		return nil, err
	}

	visualImagePath := imagePath
	visualCleanup := func() {}
	if processedPath, cleanup, err := prepareVisualZoneImage(imagePath); err == nil {
		visualImagePath = processedPath
		visualCleanup = cleanup
	}
	defer visualCleanup()

	common := []string{
		"--oem", "1",
		"--psm", "6",
		"-c", "preserve_interword_spaces=1",
		"-c", "user_defined_dpi=300",
	}

	text, err := p.runTesseractTextWithTimeout(visualImagePath, p.lang, common, 10*time.Second)
	if err != nil {
		return nil, err
	}

	vz := extractVisualZoneFields(text, 0)
	vz.RawText = text
	return vz, nil
}

func (p *OCRProcessor) ExtractText(imagePath string) (string, error) {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return "", os.ErrInvalid
	}
	if _, err := os.Stat(imagePath); err != nil {
		return "", err
	}

	common := []string{
		"--oem", "1",
		"-c", "preserve_interword_spaces=1",
		"-c", "user_defined_dpi=300",
	}

	// Run PSM 6 and PSM 11 concurrently — both are needed regardless, so
	// parallelising halves the wall-clock time on multi-core and keeps parity
	// on single-core (each Tesseract is already limited to 1 thread via OMP_THREAD_LIMIT=1).
	var (
		textPSM6, textPSM11 string
		err6, err11         error
		wg                  sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		textPSM6, err6 = p.runTesseractText(imagePath, p.lang, append([]string{"--psm", "6"}, common...))
	}()
	go func() {
		defer wg.Done()
		textPSM11, err11 = p.runTesseractText(imagePath, p.lang, append([]string{"--psm", "11"}, common...))
	}()
	wg.Wait()

	if err6 != nil && err11 != nil {
		return "", err6
	}

	return mergeOCRTextLines(textPSM6, textPSM11), nil
}

func (p *OCRProcessor) ExtractFastText(imagePath string) (string, error) {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return "", os.ErrInvalid
	}
	if _, err := os.Stat(imagePath); err != nil {
		return "", err
	}

	common := []string{
		"--oem", "1",
		"--psm", "6",
		"-c", "preserve_interword_spaces=1",
		"-c", "user_defined_dpi=300",
	}

	return p.runTesseractTextWithTimeout(imagePath, p.lang, common, 5*time.Second)
}

func (p *OCRProcessor) ExtractPassportData(imagePath string) (*PassportData, error) {
	return p.extractPassportData(imagePath, true)
}

func (p *OCRProcessor) ExtractPassportPreviewData(imagePath string) (*PassportData, error) {
	visualZoneResult := p.startVisualZoneExtraction(imagePath)

	data, err := p.extractPassportCore(imagePath)
	if err != nil {
		return nil, err
	}

	if vz := <-visualZoneResult; vz.err == nil && vz.data != nil {
		applyVisualZoneData(data, vz.data)
	}

	return data, nil
}

func (p *OCRProcessor) extractPassportData(imagePath string, includeVisualZone bool) (*PassportData, error) {
	if !includeVisualZone {
		data, err := p.extractPassportCore(imagePath)
		if err != nil {
			return nil, err
		}
		return postCorrectPassportData(data), nil
	}

	visualZoneResult := p.startVisualZoneExtraction(imagePath)

	data, err := p.extractPassportCore(imagePath)
	if err != nil {
		return nil, err
	}

	if vz := <-visualZoneResult; vz.err == nil && vz.data != nil {
		applyVisualZoneData(data, vz.data)
	}

	data.CountryCode = strings.ToUpper(strings.TrimSpace(data.CountryCode))
	data.Nationality = strings.ToUpper(strings.TrimSpace(data.Nationality))
	return postCorrectPassportData(data), nil
}

func (p *OCRProcessor) extractPassportCore(imagePath string) (*PassportData, error) {
	mrz1, mrz2, conf, err := p.ExtractMRZ(imagePath)
	if err != nil {
		return nil, err
	}

	parsed, err := p.parser.ParseMRZ(mrz1, mrz2)
	if err != nil {
		return nil, err
	}

	passportNumber := strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(parsed.PassportNumber)), "<", "")
	passportNumber = strings.ReplaceAll(passportNumber, " ", "")

	data := &PassportData{
		DocumentType:   strings.TrimSpace(parsed.DocumentType),
		CountryCode:    strings.TrimSpace(parsed.IssuingCountry),
		Surname:        strings.TrimSpace(parsed.Surname),
		GivenNames:     strings.TrimSpace(strings.Join(parsed.GivenNames, " ")),
		PassportNumber: passportNumber,
		Nationality:    strings.TrimSpace(parsed.Nationality),
		DateOfBirth:    parsed.DateOfBirth,
		Sex:            strings.TrimSpace(parsed.Sex),
		DateOfExpiry:   parsed.DateOfExpiry,
		MRZLine1:       parsed.RawLine1,
		MRZLine2:       parsed.RawLine2,
		Confidence:     conf,
		ExtractedAt:    time.Now().UTC(),
	}

	data.CountryCode = strings.ToUpper(strings.TrimSpace(data.CountryCode))
	data.Nationality = strings.ToUpper(strings.TrimSpace(data.Nationality))

	if strings.TrimSpace(data.PlaceOfBirth) == "" {
		if rawText, err := p.ExtractText(imagePath); err == nil && strings.TrimSpace(rawText) != "" {
			if pb := extractPlaceOfBirthFromRawText(rawText, data.DateOfBirth); pb != "" {
				data.PlaceOfBirth = pb
			}
		}
	}

	return data, nil
}

func (p *OCRProcessor) startVisualZoneExtraction(imagePath string) <-chan visualZoneExtractionResult {
	ch := make(chan visualZoneExtractionResult, 1)
	go func() {
		vz, err := p.ExtractVisualZone(imagePath)
		ch <- visualZoneExtractionResult{
			data: vz,
			err:  err,
		}
	}()
	return ch
}

func applyVisualZoneData(data *PassportData, vz *VisualZoneData) {
	if data == nil || vz == nil {
		return
	}

	if strings.TrimSpace(data.PlaceOfBirth) == "" && !data.DateOfBirth.IsZero() && strings.TrimSpace(vz.RawText) != "" {
		if fallbackPlace := findPlaceOfBirthNearBirthDate(vz.RawText, data.DateOfBirth); fallbackPlace != "" {
			data.PlaceOfBirth = fallbackPlace
		}
	}
	if strings.TrimSpace(data.PlaceOfBirth) == "" && strings.TrimSpace(vz.PlaceOfBirth) != "" {
		data.PlaceOfBirth = strings.TrimSpace(vz.PlaceOfBirth)
	}
	if data.DateOfIssue.IsZero() && !vz.DateOfIssue.IsZero() {
		data.DateOfIssue = vz.DateOfIssue.UTC()
	}
	if strings.TrimSpace(data.IssuingAuthority) == "" && strings.TrimSpace(vz.Authority) != "" {
		data.IssuingAuthority = strings.TrimSpace(vz.Authority)
	}
}

func uniqueOCRLanguages(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		language := strings.TrimSpace(strings.ToLower(value))
		if language == "" || slices.Contains(out, language) {
			continue
		}
		out = append(out, language)
	}
	return out
}
