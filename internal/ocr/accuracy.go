package ocr

import "strings"

type FieldAccuracy struct {
	FieldName string
	Correct   int
	Incorrect int
	Total     int
}

type AccuracyReport struct {
	Fields          []FieldAccuracy
	TotalCorrect    int
	TotalFields     int
	OverallAccuracy float64
}

func CompareFields(actual, expected *MRZData) map[string]bool {
	result := make(map[string]bool)

	if actual == nil || expected == nil {
		return result
	}

	if expected.DocumentType != "" {
		result["DocumentType"] = actual.DocumentType == expected.DocumentType
	}
	if expected.IssuingCountry != "" {
		result["IssuingCountry"] = actual.IssuingCountry == expected.IssuingCountry
	}
	if expected.Surname != "" {
		result["Surname"] = actual.Surname == expected.Surname
	}
	if len(expected.GivenNames) > 0 {
		actualStr := strings.Join(actual.GivenNames, " ")
		expectedStr := strings.Join(expected.GivenNames, " ")
		result["GivenNames"] = actualStr == expectedStr
	}
	if expected.PassportNumber != "" {
		result["PassportNumber"] = actual.PassportNumber == expected.PassportNumber
	}
	if expected.Nationality != "" {
		result["Nationality"] = actual.Nationality == expected.Nationality
	}
	if expected.Sex != "" {
		result["Sex"] = actual.Sex == expected.Sex
	}
	if !expected.DateOfBirth.IsZero() {
		result["DateOfBirth"] = actual.DateOfBirth.Equal(expected.DateOfBirth)
	}
	if !expected.DateOfExpiry.IsZero() {
		result["DateOfExpiry"] = actual.DateOfExpiry.Equal(expected.DateOfExpiry)
	}

	return result
}

func CompareFusedFields(actual *FusedMRZResult, expected *MRZData) map[string]bool {
	result := make(map[string]bool)

	if actual == nil || expected == nil {
		return result
	}

	if expected.DocumentType != "" {
		result["DocumentType"] = actual.DocumentType == expected.DocumentType
	}
	if expected.IssuingCountry != "" {
		result["IssuingCountry"] = actual.IssuingCountry == expected.IssuingCountry
	}
	if expected.Surname != "" {
		result["Surname"] = actual.Surname == expected.Surname
	}
	if len(expected.GivenNames) > 0 {
		actualStr := strings.Join(actual.GivenNames, " ")
		expectedStr := strings.Join(expected.GivenNames, " ")
		result["GivenNames"] = actualStr == expectedStr
	}
	if expected.PassportNumber != "" {
		result["PassportNumber"] = actual.PassportNumber == expected.PassportNumber
	}
	if expected.Nationality != "" {
		result["Nationality"] = actual.Nationality == expected.Nationality
	}
	if expected.Sex != "" {
		result["Sex"] = actual.Sex == expected.Sex
	}
	if !expected.DateOfBirth.IsZero() {
		result["DateOfBirth"] = actual.DateOfBirth.Equal(expected.DateOfBirth)
	}
	if !expected.DateOfExpiry.IsZero() {
		result["DateOfExpiry"] = actual.DateOfExpiry.Equal(expected.DateOfExpiry)
	}

	return result
}

func BuildAccuracyReport(results []map[string]bool) AccuracyReport {
	report := AccuracyReport{}
	totals := make(map[string]int)
	correct := make(map[string]int)

	for _, result := range results {
		for field, ok := range result {
			totals[field]++
			if ok {
				correct[field]++
			}
		}
	}

	for field, total := range totals {
		c := correct[field]
		report.Fields = append(report.Fields, FieldAccuracy{
			FieldName: field,
			Correct:   c,
			Incorrect: total - c,
			Total:     total,
		})
		report.TotalCorrect += c
		report.TotalFields += total
	}

	if report.TotalFields > 0 {
		report.OverallAccuracy = float64(report.TotalCorrect) / float64(report.TotalFields)
	}

	return report
}
