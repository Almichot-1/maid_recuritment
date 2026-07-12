package ocr

import (
	"time"
)

func rerunLowConfidenceFields(p *OCRProcessor, imagePath string, fused *FusedMRZResult) *FusedMRZResult {
	if fused == nil || fused.SuccessfulPasses == 0 {
		return fused
	}

	lowFields := findLowConfidenceFields(fused.FieldConfidence)
	if len(lowFields) == 0 || fused.Confidence >= 0.9 {
		return fused
	}

	languages := getLanguages(p)
	rerunCfg := MRZPassConfig{PSMMode: 3, OEMMode: 1, Label: "rerun_oem1_psm3"}

	for _, lang := range languages {
		result := runSingleMRZPass(p, imagePath, lang, rerunCfg, 7*time.Second)
		if result.Err != nil || result.Parsed == nil {
			continue
		}
		if result.Parsed.IsValid && result.Confidence > fused.Confidence {
			if result.Line1 != "" && result.Line2 != "" {
				fused.Line1 = result.Line1
				fused.Line2 = result.Line2
			}
			if result.Parsed.PassportNumber != "" {
				fused.PassportNumber = result.Parsed.PassportNumber
			}
			if result.Parsed.Surname != "" {
				fused.Surname = result.Parsed.Surname
			}
			if len(result.Parsed.GivenNames) > 0 {
				fused.GivenNames = result.Parsed.GivenNames
			}
			if !result.Parsed.DateOfBirth.IsZero() {
				fused.DateOfBirth = result.Parsed.DateOfBirth
			}
			if !result.Parsed.DateOfExpiry.IsZero() {
				fused.DateOfExpiry = result.Parsed.DateOfExpiry
			}
			fused.Confidence = result.Confidence
			fused.SuccessfulPasses++
			break
		}
	}

	return fused
}

func findLowConfidenceFields(fc map[string]float64) []string {
	var fields []string
	threshold := 0.8
	for field, conf := range fc {
		if conf < threshold {
			fields = append(fields, field)
		}
	}
	return fields
}

func getLanguages(p *OCRProcessor) []string {
	langs := []string{"eng"}
	if p.lang != "" && p.lang != "eng" {
		langs = append(langs, p.lang)
	}
	return uniqueOCRLanguages(langs...)
}
