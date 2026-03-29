package ocr

import (
	"fmt"
	"os"
	"slices"
	"strings"
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

	mrzImagePath := imagePath
	mrzCleanup := func() {}
	if processedPath, cleanup, err := prepareMRZImage(imagePath); err == nil {
		mrzImagePath = processedPath
		mrzCleanup = cleanup
	}
	defer mrzCleanup()

	mrzArgs := []string{
		"--psm", "6",
		"-c", "tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<",
	}
	languages := uniqueOCRLanguages("ocrb", "eng", p.lang)
	attemptErrors := make([]string, 0, len(languages))

	for _, language := range languages {
		text, err := p.runTesseractText(mrzImagePath, language, mrzArgs)
		if err != nil {
			attemptErrors = append(attemptErrors, fmt.Sprintf("%s: %v", language, err))
			continue
		}

		line1, line2 := extractMRZLinesFromText(text)
		if line1 != "" && line2 != "" {
			return line1, line2, 0, nil
		}

		attemptErrors = append(attemptErrors, fmt.Sprintf("%s: invalid MRZ output", language))
	}

	if len(attemptErrors) == 0 {
		return "", "", 0, ErrInvalidMRZOutput
	}

	return "", "", 0, fmt.Errorf("%w: %s", ErrInvalidMRZOutput, strings.Join(attemptErrors, "; "))
}

func (p *OCRProcessor) ExtractVisualZone(imagePath string) (*VisualZoneData, error) {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return nil, os.ErrInvalid
	}
	if _, err := os.Stat(imagePath); err != nil {
		return nil, err
	}

	common := []string{
		"--oem", "1",
		"-c", "preserve_interword_spaces=1",
		"-c", "user_defined_dpi=300",
	}

	textPSM6, err6 := p.runTesseractText(imagePath, p.lang, append([]string{"--psm", "6"}, common...))
	textPSM11, err11 := p.runTesseractText(imagePath, p.lang, append([]string{"--psm", "11"}, common...))
	if err6 != nil && err11 != nil {
		return nil, err6
	}
	text := mergeOCRTextLines(textPSM6, textPSM11)

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

	textPSM6, err6 := p.runTesseractText(imagePath, p.lang, append([]string{"--psm", "6"}, common...))
	textPSM11, err11 := p.runTesseractText(imagePath, p.lang, append([]string{"--psm", "11"}, common...))
	if err6 != nil && err11 != nil {
		return "", err6
	}

	return mergeOCRTextLines(textPSM6, textPSM11), nil
}

func (p *OCRProcessor) ExtractPassportData(imagePath string) (*PassportData, error) {
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

	if vz, err := p.ExtractVisualZone(imagePath); err == nil && vz != nil {
		if strings.TrimSpace(vz.PlaceOfBirth) != "" {
			data.PlaceOfBirth = strings.TrimSpace(vz.PlaceOfBirth)
		}
		if !vz.DateOfIssue.IsZero() {
			data.DateOfIssue = vz.DateOfIssue.UTC()
		}
		if strings.TrimSpace(vz.Authority) != "" {
			data.IssuingAuthority = strings.TrimSpace(vz.Authority)
		}
	}

	data.CountryCode = strings.ToUpper(strings.TrimSpace(data.CountryCode))
	data.Nationality = strings.ToUpper(strings.TrimSpace(data.Nationality))
	return data, nil
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
