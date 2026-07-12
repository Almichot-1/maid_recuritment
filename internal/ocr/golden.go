package ocr

import (
	"fmt"
	"strings"
)

// GoldenCase defines an expected MRZ parse result.
// Used for regression testing — input goes through ParseMRZ or the full pipeline,
// and the output is compared against Expected fields.
type GoldenCase struct {
	Name                string
	Line1               string
	Line2               string
	ExpectedDocType     string
	ExpectedCountry     string
	ExpectedSurname     string
	ExpectedGivenNames  []string
	ExpectedPassport    string
	ExpectedNationality string
	ExpectedSex         string
	ExpectedIsValid     bool
	ExpectedConfidence  float64
	ConfidenceTolerance float64
	ExpectedError       bool
	ExpectedCorrections int
}

// NewGoldenCase creates a GoldenCase with standard tolerance.
func NewGoldenCase(name, line1, line2 string) GoldenCase {
	return GoldenCase{
		Name:                name,
		Line1:               line1,
		Line2:               line2,
		ConfidenceTolerance: 0.05,
	}
}

// RunGoldenTest runs a single golden test through ParseMRZ and validates
// the output against expected fields. Returns an error message or empty string.
func RunGoldenTest(gc GoldenCase) string {
	parser := NewMRZParser()
	data, err := parser.ParseMRZ(gc.Line1, gc.Line2)

	if gc.ExpectedError {
		if err == nil {
			return "expected error but got none"
		}
		return ""
	}
	if err != nil {
		return "unexpected error: " + err.Error()
	}
	if data == nil {
		return "data is nil"
	}

	var failures []string

	if gc.ExpectedDocType != "" {
		checkField("DocumentType", data.DocumentType, gc.ExpectedDocType, &failures)
	}
	if gc.ExpectedCountry != "" {
		checkField("IssuingCountry", data.IssuingCountry, gc.ExpectedCountry, &failures)
	}
	if gc.ExpectedSurname != "" {
		checkField("Surname", data.Surname, gc.ExpectedSurname, &failures)
	}
	if gc.ExpectedPassport != "" {
		checkField("PassportNumber", data.PassportNumber, gc.ExpectedPassport, &failures)
	}
	if gc.ExpectedNationality != "" {
		checkField("Nationality", data.Nationality, gc.ExpectedNationality, &failures)
	}
	if gc.ExpectedSex != "" {
		checkField("Sex", data.Sex, gc.ExpectedSex, &failures)
	}

	if len(gc.ExpectedGivenNames) > 0 {
		givenStr := strings.Join(data.GivenNames, " ")
		expectedGivenStr := strings.Join(gc.ExpectedGivenNames, " ")
		if givenStr != expectedGivenStr {
			failures = append(failures, "GivenNames: got "+givenStr+", want "+expectedGivenStr)
		}
	}

	if gc.ExpectedIsValid != data.IsValid {
		failures = append(failures, "IsValid: got "+fmt.Sprint(data.IsValid)+", want "+fmt.Sprint(gc.ExpectedIsValid))
	}

	if gc.ExpectedConfidence > 0 {
		diff := gc.ExpectedConfidence - data.Confidence
		if diff < 0 {
			diff = -diff
		}
		if diff > gc.ConfidenceTolerance {
			failures = append(failures, fmt.Sprintf("Confidence: got %.2f, want %.2f (tolerance %.2f)", data.Confidence, gc.ExpectedConfidence, gc.ConfidenceTolerance))
		}
	}

	if data.CorrectionCount != gc.ExpectedCorrections {
		failures = append(failures, fmt.Sprintf("CorrectionCount: got %d, want %d", data.CorrectionCount, gc.ExpectedCorrections))
	}

	if len(failures) > 0 {
		return gc.Name + ": " + strings.Join(failures, "; ")
	}
	return ""
}

func checkField(name, got, want string, failures *[]string) {
	if got != want {
		*failures = append(*failures, fmt.Sprintf("%s: got %q, want %q", name, got, want))
	}
}
