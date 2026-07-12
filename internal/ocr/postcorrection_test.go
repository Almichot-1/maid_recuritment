package ocr

import (
	"testing"
	"time"
)

func testDate(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func TestIsValidCountryCode_Valid(t *testing.T) {
	for _, code := range []string{"USA", "GBR", "JPN", "DEU"} {
		if !isValidCountryCode(code) {
			t.Fatalf("expected %q to be valid", code)
		}
	}
}

func TestIsValidCountryCode_Invalid(t *testing.T) {
	for _, code := range []string{"XYZ", "A12", ""} {
		if isValidCountryCode(code) {
			t.Fatalf("expected %q to be invalid", code)
		}
	}
}

func TestIsValidCountryCode_Lowercase(t *testing.T) {
	if isValidCountryCode("usa") {
		t.Fatal("expected lowercase 'usa' to be invalid")
	}
}

func TestCorrectCountryCode_AlreadyValid(t *testing.T) {
	result, changed := correctCountryCode("USA")
	if changed {
		t.Fatal("expected no change for already valid code")
	}
	if result != "USA" {
		t.Fatalf("expected 'USA', got %q", result)
	}
}

func TestCorrectCountryCode_Invalid(t *testing.T) {
	result, changed := correctCountryCode("ZZZ")
	if changed {
		t.Fatal("expected no change for unresolvable code")
	}
	if result != "ZZZ" {
		t.Fatalf("expected 'ZZZ', got %q", result)
	}
}

func TestCorrectCountryCode_Empty(t *testing.T) {
	result, changed := correctCountryCode("")
	if changed {
		t.Fatal("expected no change for empty code")
	}
	if result != "" {
		t.Fatalf("expected empty, got %q", result)
	}
}

func TestCorrectNationality_Valid(t *testing.T) {
	result, changed := correctNationality("USA")
	if changed {
		t.Fatal("expected no change for valid nationality")
	}
	if result != "USA" {
		t.Fatalf("expected 'USA', got %q", result)
	}
}

func TestGetISOCountryCodes_NotEmpty(t *testing.T) {
	codes := GetISOCountryCodes()
	if len(codes) == 0 {
		t.Fatal("expected non-empty country code map")
	}
}

func TestGetISOCountryCodes_ContainsUSA(t *testing.T) {
	codes := GetISOCountryCodes()
	if name, ok := codes["USA"]; !ok {
		t.Fatal("expected map to contain USA")
	} else if name != "UNITED STATES" {
		t.Fatalf("expected 'UNITED STATES', got %q", name)
	}
}

func TestValidateDateConsistency_AllValid(t *testing.T) {
	dob := testDate(1990, 1, 1)
	issue := testDate(2010, 6, 15)
	expiry := testDate(2030, 6, 14)
	warnings := validateDateConsistency(dob, issue, expiry)
	if warnings != nil {
		t.Fatalf("expected nil, got %v", warnings)
	}
}

func TestValidateDateConsistency_IssueBeforeBirth(t *testing.T) {
	dob := testDate(1990, 1, 1)
	issue := testDate(1985, 1, 1)
	expiry := testDate(2005, 1, 1)
	warnings := validateDateConsistency(dob, issue, expiry)
	if len(warnings) < 1 {
		t.Fatal("expected at least one warning for issue before birth")
	}
}

func TestValidateDateConsistency_ExpiryBeforeIssue(t *testing.T) {
	dob := testDate(1990, 1, 1)
	issue := testDate(2010, 1, 1)
	expiry := testDate(2005, 1, 1)
	warnings := validateDateConsistency(dob, issue, expiry)
	if len(warnings) < 1 {
		t.Fatal("expected at least one warning for expiry before issue")
	}
}

func TestValidateDateConsistency_BirthTooOld(t *testing.T) {
	dob := testDate(1850, 1, 1)
	issue := testDate(1900, 1, 1)
	expiry := testDate(1920, 1, 1)
	warnings := validateDateConsistency(dob, issue, expiry)
	if len(warnings) < 1 {
		t.Fatal("expected at least one warning for birth too old")
	}
}

func TestValidateDateConsistency_ZeroDates(t *testing.T) {
	var zero time.Time
	warnings := validateDateConsistency(zero, zero, zero)
	if warnings != nil {
		t.Fatalf("expected nil for all zero dates, got %v", warnings)
	}
}

func TestValidateNames_Valid(t *testing.T) {
	warnings := validateNames("SMITH", []string{"JOHN"})
	if warnings != nil {
		t.Fatalf("expected nil, got %v", warnings)
	}
}

func TestValidateNames_EmptySurname(t *testing.T) {
	warnings := validateNames("", []string{"JOHN"})
	if len(warnings) < 1 {
		t.Fatal("expected at least one warning for empty surname")
	}
}

func TestValidateNames_SingleCharSurname(t *testing.T) {
	warnings := validateNames("S", []string{"JOHN"})
	if len(warnings) < 1 {
		t.Fatal("expected at least one warning for single char surname")
	}
}

func TestValidateNames_SingleGivenNameInitial(t *testing.T) {
	warnings := validateNames("SMITH", []string{"J"})
	if len(warnings) < 1 {
		t.Fatal("expected at least one warning for single char given name")
	}
}

func TestValidateNames_AllSameChar(t *testing.T) {
	warnings := validateNames("AAAA", []string{"JOHN"})
	if len(warnings) < 1 {
		t.Fatal("expected at least one warning for all-same-char surname")
	}
}

func TestValidateNames_EmptyGiven(t *testing.T) {
	warnings := validateNames("SMITH", nil)
	if len(warnings) < 1 {
		t.Fatal("expected at least one warning for empty given names")
	}
}

func TestCorrectNameOcr_DigitZero(t *testing.T) {
	result, changed := correctNameOcr("J0HN")
	if !changed {
		t.Fatal("expected change for 'J0HN'")
	}
	if result != "JOHN" {
		t.Fatalf("expected 'JOHN', got %q", result)
	}
}

func TestCorrectNameOcr_DigitOne(t *testing.T) {
	result, changed := correctNameOcr("AR1EL")
	if !changed {
		t.Fatal("expected change for 'AR1EL'")
	}
	if result != "ARIEL" {
		t.Fatalf("expected 'ARIEL', got %q", result)
	}
}

func TestCorrectNameOcr_NoChange(t *testing.T) {
	result, changed := correctNameOcr("JOHN")
	if changed {
		t.Fatal("expected no change for 'JOHN'")
	}
	if result != "JOHN" {
		t.Fatalf("expected 'JOHN', got %q", result)
	}
}

func TestCorrectNameOcr_Empty(t *testing.T) {
	result, changed := correctNameOcr("")
	if changed {
		t.Fatal("expected no change for empty string")
	}
	if result != "" {
		t.Fatalf("expected empty, got %q", result)
	}
}

func TestCorrectNameOcr_Multiple(t *testing.T) {
	result, changed := correctNameOcr("M8R1A")
	if !changed {
		t.Fatal("expected change for 'M8R1A'")
	}
	if result != "MBRIA" {
		t.Fatalf("expected 'MBRIA', got %q", result)
	}
}

func TestIsAllSameCharacter_AllSame(t *testing.T) {
	if !isAllSameCharacter("AAA") {
		t.Fatal("expected true for 'AAA'")
	}
}

func TestIsAllSameCharacter_Different(t *testing.T) {
	if isAllSameCharacter("ABC") {
		t.Fatal("expected false for 'ABC'")
	}
}

func TestIsAllSameCharacter_Empty(t *testing.T) {
	if isAllSameCharacter("") {
		t.Fatal("expected false for empty string")
	}
}

func TestPostCorrectPassportData_Nil(t *testing.T) {
	result := postCorrectPassportData(nil)
	if result != nil {
		t.Fatal("expected nil for nil input")
	}
}

func TestPostCorrectPassportData_NameCorrection(t *testing.T) {
	data := &PassportData{
		Surname:    "SM1TH",
		GivenNames: "J0HN",
		Confidence: 1.0,
	}
	result := postCorrectPassportData(data)
	if result.Surname != "SMITH" {
		t.Fatalf("expected surname 'SMITH', got %q", result.Surname)
	}
	if result.GivenNames != "JOHN" {
		t.Fatalf("expected given names 'JOHN', got %q", result.GivenNames)
	}
}

func TestPostCorrectPassportData_InvalidCountryPenalty(t *testing.T) {
	data := &PassportData{
		CountryCode: "XYZ",
		Nationality: "USA",
		Confidence:  1.0,
	}
	result := postCorrectPassportData(data)
	if result.Confidence >= 1.0 {
		t.Fatal("expected confidence penalty for invalid country code")
	}
}
