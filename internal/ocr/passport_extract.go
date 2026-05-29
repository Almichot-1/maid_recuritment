package ocr

import (
	"fmt"
	"strings"
	"time"
)

// PassportExtractResult is the structured output of passport OCR for services.
type PassportExtractResult struct {
	GivenNames       string
	Surname          string
	PassportNumber   string
	CountryCode      string
	Nationality      string
	DateOfBirth      time.Time
	DateOfExpiry     time.Time
	DateOfIssue      time.Time
	PlaceOfBirth     string
	Sex              string
	IssuingAuthority string
	MRZLine1         string
	MRZLine2         string
	Confidence       float64
}

// ExtractPassportData runs MRZ and visual-zone OCR on a passport image.
func (p *OCRProcessor) ExtractPassportData(imagePath string) (*PassportExtractResult, error) {
	line1, line2, confidence, err := p.ExtractMRZ(imagePath)
	if err != nil {
		return nil, err
	}

	mrz, err := p.parser.ParseMRZ(line1, line2)
	if err != nil {
		return nil, fmt.Errorf("parse MRZ: %w", err)
	}

	result := &PassportExtractResult{
		GivenNames:     strings.Join(mrz.GivenNames, " "),
		Surname:        mrz.Surname,
		PassportNumber: strings.Trim(mrz.PassportNumber, "<"),
		CountryCode:    mrz.IssuingCountry,
		Nationality:    mrz.Nationality,
		DateOfBirth:    mrz.DateOfBirth,
		DateOfExpiry:   mrz.DateOfExpiry,
		Sex:            mrz.Sex,
		MRZLine1:       line1,
		MRZLine2:       line2,
		Confidence:     confidence,
	}

	if visual, vzErr := p.ExtractVisualZone(imagePath); vzErr == nil && visual != nil {
		result.PlaceOfBirth = visual.PlaceOfBirth
		result.IssuingAuthority = visual.Authority
		if !visual.DateOfIssue.IsZero() {
			result.DateOfIssue = visual.DateOfIssue
		}
	}

	return result, nil
}
