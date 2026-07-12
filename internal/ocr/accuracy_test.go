package ocr

import (
	"testing"
	"time"
)

func makeMRZData(docType, country, surname string, givenNames []string, passport, nationality, sex string) *MRZData {
	return &MRZData{
		DocumentType:   docType,
		IssuingCountry: country,
		Surname:        surname,
		GivenNames:     givenNames,
		PassportNumber: passport,
		Nationality:    nationality,
		Sex:            sex,
		IsValid:        true,
	}
}

func makeFusedResult(docType, country, surname string, givenNames []string, passport, nationality, sex string) *FusedMRZResult {
	return &FusedMRZResult{
		DocumentType:   docType,
		IssuingCountry: country,
		Surname:        surname,
		GivenNames:     givenNames,
		PassportNumber: passport,
		Nationality:    nationality,
		Sex:            sex,
	}
}

func TestCompareFields_AllMatch(t *testing.T) {
	actual := makeMRZData("P", "UTO", "SMITH", []string{"JOHN", "A"}, "AB123456", "UTO", "M")
	expected := makeMRZData("P", "UTO", "SMITH", []string{"JOHN", "A"}, "AB123456", "UTO", "M")

	results := CompareFields(actual, expected)

	for field, ok := range results {
		if !ok {
			t.Errorf("field %s should match but did not", field)
		}
	}

	if len(results) != 7 {
		t.Errorf("expected 7 fields compared, got %d", len(results))
	}
}

func TestCompareFields_SomeMismatch(t *testing.T) {
	actual := makeMRZData("P", "UTO", "SMITH", []string{"JOHN"}, "AB123456", "UTO", "M")
	expected := makeMRZData("P", "UTO", "JONES", []string{"JOHN"}, "XY999999", "UTO", "F")

	results := CompareFields(actual, expected)

	if results["Surname"] {
		t.Error("Surname should mismatch")
	}
	if results["PassportNumber"] {
		t.Error("PassportNumber should mismatch")
	}
	if results["Sex"] {
		t.Error("Sex should mismatch")
	}
	if !results["DocumentType"] {
		t.Error("DocumentType should match")
	}
	if !results["IssuingCountry"] {
		t.Error("IssuingCountry should match")
	}
	if !results["Nationality"] {
		t.Error("Nationality should match")
	}
	if !results["GivenNames"] {
		t.Error("GivenNames should match")
	}
}

func TestCompareFields_NilInput(t *testing.T) {
	results := CompareFields(nil, nil)
	if len(results) != 0 {
		t.Errorf("expected empty map for nil inputs, got %d entries", len(results))
	}

	actual := makeMRZData("P", "UTO", "SMITH", nil, "AB123456", "UTO", "M")
	results = CompareFields(actual, nil)
	if len(results) != 0 {
		t.Errorf("expected empty map when expected is nil, got %d entries", len(results))
	}

	results = CompareFields(nil, actual)
	if len(results) != 0 {
		t.Errorf("expected empty map when actual is nil, got %d entries", len(results))
	}
}

func TestCompareFields_EmptyExpected(t *testing.T) {
	actual := makeMRZData("P", "UTO", "SMITH", []string{"JOHN"}, "AB123456", "UTO", "M")
	expected := &MRZData{}

	results := CompareFields(actual, expected)
	if len(results) != 0 {
		t.Errorf("expected empty map for empty expected fields, got %d entries", len(results))
	}
}

func TestCompareFusedFields_Match(t *testing.T) {
	actual := makeFusedResult("P", "UTO", "SMITH", []string{"JOHN"}, "AB123456", "UTO", "M")
	expected := makeMRZData("P", "UTO", "SMITH", []string{"JOHN"}, "AB123456", "UTO", "M")

	results := CompareFusedFields(actual, expected)

	for field, ok := range results {
		if !ok {
			t.Errorf("field %s should match but did not", field)
		}
	}

	if len(results) != 7 {
		t.Errorf("expected 7 fields compared, got %d", len(results))
	}
}

func TestBuildAccuracyReport_AllCorrect(t *testing.T) {
	r1 := map[string]bool{"DocumentType": true, "Surname": true, "Sex": true}
	r2 := map[string]bool{"DocumentType": true, "Surname": true, "Sex": true}

	report := BuildAccuracyReport([]map[string]bool{r1, r2})

	if report.OverallAccuracy != 1.0 {
		t.Errorf("expected accuracy 1.0, got %f", report.OverallAccuracy)
	}
	if report.TotalCorrect != 6 {
		t.Errorf("expected 6 total correct, got %d", report.TotalCorrect)
	}
	if report.TotalFields != 6 {
		t.Errorf("expected 6 total fields, got %d", report.TotalFields)
	}
}

func TestBuildAccuracyReport_Mixed(t *testing.T) {
	r1 := map[string]bool{"DocumentType": true, "Surname": true, "Sex": true}
	r2 := map[string]bool{"DocumentType": true, "Surname": false, "Sex": true}
	r3 := map[string]bool{"DocumentType": false, "Surname": true, "Sex": false}

	report := BuildAccuracyReport([]map[string]bool{r1, r2, r3})

	if report.TotalCorrect != 6 {
		t.Errorf("expected 6 total correct, got %d", report.TotalCorrect)
	}
	if report.TotalFields != 9 {
		t.Errorf("expected 9 total fields, got %d", report.TotalFields)
	}
	if report.OverallAccuracy != 6.0/9.0 {
		t.Errorf("expected accuracy %f, got %f", 6.0/9.0, report.OverallAccuracy)
	}

	if len(report.Fields) != 3 {
		t.Errorf("expected 3 field entries, got %d", len(report.Fields))
	}
}

func TestBuildAccuracyReport_Empty(t *testing.T) {
	report := BuildAccuracyReport([]map[string]bool{})

	if report.OverallAccuracy != 0.0 {
		t.Errorf("expected accuracy 0.0 for empty input, got %f", report.OverallAccuracy)
	}
	if report.TotalCorrect != 0 {
		t.Errorf("expected 0 total correct, got %d", report.TotalCorrect)
	}
	if report.TotalFields != 0 {
		t.Errorf("expected 0 total fields, got %d", report.TotalFields)
	}
	if len(report.Fields) != 0 {
		t.Errorf("expected 0 field entries, got %d", len(report.Fields))
	}
}

func TestCompareFields_DateComparison(t *testing.T) {
	dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	doe := time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)

	actual := makeMRZData("P", "UTO", "SMITH", []string{"JOHN"}, "AB123456", "UTO", "M")
	actual.DateOfBirth = dob
	actual.DateOfExpiry = doe

	expected := makeMRZData("P", "UTO", "SMITH", []string{"JOHN"}, "AB123456", "UTO", "M")
	expected.DateOfBirth = dob
	expected.DateOfExpiry = doe

	results := CompareFields(actual, expected)

	if !results["DateOfBirth"] {
		t.Error("DateOfBirth should match")
	}
	if !results["DateOfExpiry"] {
		t.Error("DateOfExpiry should match")
	}
}
