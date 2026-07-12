package ocr

import (
	"testing"
	"time"
)

// --- CleanMRZLine Tests ---

func TestCleanMRZLine_Valid(t *testing.T) {
	line, err := CleanMRZLine("P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<"
	if line != expected {
		t.Fatalf("expected %q, got %q", expected, line)
	}
}

func TestCleanMRZLine_Empty(t *testing.T) {
	_, err := CleanMRZLine("")
	if err == nil {
		t.Fatal("expected error for empty line")
	}
}

func TestCleanMRZLine_WhitespaceRemoved(t *testing.T) {
	line, err := CleanMRZLine("P< UTOERIKSSON << ANNA < MARIA <<<<<<<<<<<<<<<<<<<")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<"
	if line != expected {
		t.Fatalf("expected %q, got %q", expected, line)
	}
}

func TestCleanMRZLine_UnicodeReplacements(t *testing.T) {
	// Use the same mojibake bytes that the replacement map handles:
	// "\xc3\x82\xc2\xab" = "Â«" (U+00C2 U+00AB), a common double-encoding of «
	line, err := CleanMRZLine("P<UTO\xc3\x82\xc2\xabERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "P<UTO<ERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<"
	if line != expected {
		t.Fatalf("expected %q, got %q", expected, line)
	}
}

func TestCleanMRZLine_InvalidCharacters(t *testing.T) {
	_, err := CleanMRZLine("P<UTØERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<")
	if err == nil {
		t.Fatal("expected error for invalid characters")
	}
}

// --- ParseMRZ Tests ---

var (
	testLine1 = "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<"
	testLine2 = "L898902C<3UTO6908061F9406236ZE184226B<<<<<14"
)

func TestParseMRZ_Valid(t *testing.T) {
	p := NewMRZParser()
	data, err := p.ParseMRZ(testLine1, testLine2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !data.IsValid {
		t.Fatalf("expected valid MRZ, got errors: %v", data.ValidationErrors)
	}
	if data.DocumentType != "P" {
		t.Fatalf("expected document type P, got %q", data.DocumentType)
	}
	if data.IssuingCountry != "UTO" {
		t.Fatalf("expected issuing country UTO, got %q", data.IssuingCountry)
	}
	if data.Surname != "ERIKSSON" {
		t.Fatalf("expected surname ERIKSSON, got %q", data.Surname)
	}
	if len(data.GivenNames) != 2 || data.GivenNames[0] != "ANNA" || data.GivenNames[1] != "MARIA" {
		t.Fatalf("expected given names [ANNA MARIA], got %v", data.GivenNames)
	}
	if data.PassportNumber != "L898902C<" {
		t.Fatalf("expected passport L898902C<, got %q", data.PassportNumber)
	}
	if data.Nationality != "UTO" {
		t.Fatalf("expected nationality UTO, got %q", data.Nationality)
	}
	if data.Sex != "F" {
		t.Fatalf("expected sex F, got %q", data.Sex)
	}
}

func TestParseMRZ_Line1WrongLength(t *testing.T) {
	p := NewMRZParser()
	_, err := p.ParseMRZ("SHORT", testLine2)
	if err == nil {
		t.Fatal("expected error for short line1")
	}
}

func TestParseMRZ_Line2WrongLength(t *testing.T) {
	p := NewMRZParser()
	_, err := p.ParseMRZ(testLine1, "SHORT")
	if err == nil {
		t.Fatal("expected error for short line2")
	}
}

// --- Check Digit Correction Tests ---

func TestCorrectCheckDigitField_BTo8InDOB(t *testing.T) {
	// B looks like 8 in OCR fonts, confusionPairs: B→8
	// Test with DOB field "690806": B at position 3 instead of 8
	field := "690B06"
	corrected, count, err := CorrectCheckDigitField(field, '1')
	if err != nil {
		t.Fatalf("expected correction to succeed, got: %v", err)
	}
	if corrected != "690806" {
		t.Fatalf("expected 690806, got %q", corrected)
	}
	if count != 1 {
		t.Fatalf("expected 1 correction, got %d", count)
	}
}

func TestCorrectCheckDigitField_BTo8(t *testing.T) {
	// B looks like 8 in OCR fonts, confusionPairs: B→8
	// passport "L898902C" with B instead of 8 at position 3
	field := "L89B902C"
	corrected, count, err := CorrectCheckDigitField(field, '3')
	if err != nil {
		t.Fatalf("expected correction to succeed, got: %v", err)
	}
	if corrected != "L898902C" {
		t.Fatalf("expected L898902C, got %q", corrected)
	}
	if count != 1 {
		t.Fatalf("expected 1 correction, got %d", count)
	}
}

func TestCorrectCheckDigitField_AlreadyValid(t *testing.T) {
	correctField := "L898902C"
	_, _, err := CorrectCheckDigitField(correctField, '3')
	if err == nil {
		t.Fatal("expected error for already-valid field")
	}
}

func TestCorrectCheckDigitField_TooManyErrors(t *testing.T) {
	field := "L8A8A02C"
	_, _, err := CorrectCheckDigitField(field, '3')
	if err == nil {
		t.Fatal("expected error for field with too many errors")
	}
}

func TestCorrectCheckDigitField_NoConfusionPair(t *testing.T) {
	// A is not in confusionPairs, so should fail to correct
	field := "LA98902C"
	_, _, err := CorrectCheckDigitField(field, '3')
	if err == nil {
		t.Fatal("expected error for field with uncorrectable character A")
	}
}

// --- ParseMRZ with Correction Tests ---
// CleanMRZLine normalizes O→0 and I→1 (when adjacent to digits), so CorrectionCount
// may be 0 when those are fixed during the cleaning step rather than by
// CorrectCheckDigitField.

func TestParseMRZ_WithPassportNumberAutofix(t *testing.T) {
	// O should be 0 in passport number. CleanMRZLine fixes this during cleaning.
	confusedLine2 := "L8989O2C<3UTO6908061F9406236ZE184226B<<<<<14"
	p := NewMRZParser()
	data, err := p.ParseMRZ(testLine1, confusedLine2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !data.IsValid {
		t.Fatalf("expected MRZ to be valid, got errors: %v", data.ValidationErrors)
	}
	if data.PassportNumber != "L898902C<" {
		t.Fatalf("expected passport L898902C<, got %q", data.PassportNumber)
	}
}

func TestParseMRZ_Uncorrectable(t *testing.T) {
	// Multiple errors that single-char correction can't fix
	badLine2 := "LA98902C<3UTO6908061F9406236ZE184226B<<<<<14"
	p := NewMRZParser()
	data, err := p.ParseMRZ(testLine1, badLine2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.IsValid {
		t.Fatal("expected MRZ to be invalid (uncorrectable errors)")
	}
	if len(data.ValidationErrors) == 0 {
		t.Fatal("expected validation errors")
	}
}

// --- Confidence Scoring Tests ---

func TestComputeConfidence_Valid(t *testing.T) {
	p := NewMRZParser()
	data, _ := p.ParseMRZ(testLine1, testLine2)
	conf := ComputeConfidence(data)
	if conf != 1.0 {
		t.Fatalf("expected confidence 1.0 for valid MRZ, got %f", conf)
	}
}

func TestComputeConfidence_PartialCorrection(t *testing.T) {
	data := &MRZData{
		CorrectionCount: 1,
		ValidationErrors: []string{"passport number check digit corrected"},
	}
	conf := ComputeConfidence(data)
	// 1 error + 1 correction penalty: confidence = (5-1)/5 - 1*0.1 = 0.8 - 0.1 = 0.7
	if conf < 0.69 || conf > 0.71 {
		t.Fatalf("expected confidence ~0.7, got %f", conf)
	}
}

func TestComputeConfidence_NilInput(t *testing.T) {
	conf := ComputeConfidence(nil)
	if conf != 0 {
		t.Fatalf("expected 0 for nil input, got %f", conf)
	}
}

// --- MRZ Date Parsing Tests ---

func TestParseMRZBirthDate_PastCentury(t *testing.T) {
	asOf := time.Date(2026, time.March, 28, 0, 0, 0, 0, time.UTC)
	value, err := ParseMRZBirthDate("990213", asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(1999, time.February, 13, 0, 0, 0, 0, time.UTC)
	if !value.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected.Format(time.DateOnly), value.Format(time.DateOnly))
	}
}

func TestParseMRZBirthDate_CurrentCentury(t *testing.T) {
	asOf := time.Date(2026, time.March, 28, 0, 0, 0, 0, time.UTC)
	value, err := ParseMRZBirthDate("050101", asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2005, time.January, 1, 0, 0, 0, 0, time.UTC)
	if !value.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected.Format(time.DateOnly), value.Format(time.DateOnly))
	}
}

func TestParseMRZBirthDate_InvalidMonth(t *testing.T) {
	_, err := ParseMRZBirthDate("991301", time.Now().UTC())
	if err == nil {
		t.Fatal("expected error for invalid month 13")
	}
}

func TestParseMRZExpiryDate_SelectsNearest(t *testing.T) {
	asOf := time.Date(2026, time.March, 28, 0, 0, 0, 0, time.UTC)
	value, err := ParseMRZExpiryDate("290325", asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2029, time.March, 25, 0, 0, 0, 0, time.UTC)
	if !value.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected.Format(time.DateOnly), value.Format(time.DateOnly))
	}
}

func TestParseMRZExpiryDate_LeapYear(t *testing.T) {
	asOf := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	value, err := ParseMRZExpiryDate("240229", asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC)
	if !value.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected.Format(time.DateOnly), value.Format(time.DateOnly))
	}
}

// --- ParseName Tests ---

func TestParseName_SingleGivenName(t *testing.T) {
	field := "MUSTERMANN<<ERIKA<<<<<<<<<<<<<<<<<<<<<<<<"
	surname, givenNames := ParseName(field)
	if surname != "MUSTERMANN" {
		t.Fatalf("expected surname MUSTERMANN, got %q", surname)
	}
	if len(givenNames) != 1 || givenNames[0] != "ERIKA" {
		t.Fatalf("expected given names [ERIKA], got %v", givenNames)
	}
}

func TestParseName_MultipleGivenNames(t *testing.T) {
	field := "MUSTERMANN<<ERIKA<MARIA<<<<<<<<<<<<<<<<<<<"
	surname, givenNames := ParseName(field)
	if surname != "MUSTERMANN" {
		t.Fatalf("expected surname MUSTERMANN, got %q", surname)
	}
	if len(givenNames) != 2 || givenNames[0] != "ERIKA" || givenNames[1] != "MARIA" {
		t.Fatalf("expected given names [ERIKA MARIA], got %v", givenNames)
	}
}

// --- ValidateCheckDigit Tests ---

func TestCalculateCheckDigit_Valid(t *testing.T) {
	result := CalculateCheckDigit("L898902C")
	if result != 3 {
		t.Fatalf("expected check digit 3 for L898902C, got %d", result)
	}
}

func TestValidateCheckDigit_Valid(t *testing.T) {
	if !ValidateCheckDigit("L898902C", '3') {
		t.Fatal("expected valid check digit")
	}
}

func TestValidateCheckDigit_Invalid(t *testing.T) {
	if ValidateCheckDigit("L898902C", '0') {
		t.Fatal("expected invalid check digit")
	}
}

func TestValidateCheckDigit_NonDigitCheckChar(t *testing.T) {
	if ValidateCheckDigit("L898902C", 'X') {
		t.Fatal("expected invalid check digit for non-digit char")
	}
}

// --- normalizeDigit Tests ---

func TestNormalizeDigit_OTo0(t *testing.T) {
	result := normalizeDigit('O')
	if result != '0' {
		t.Fatalf("expected '0', got '%c'", result)
	}
}

func TestNormalizeDigit_ITo1(t *testing.T) {
	result := normalizeDigit('I')
	if result != '1' {
		t.Fatalf("expected '1', got '%c'", result)
	}
}

func TestNormalizeDigit_AlreadyDigit(t *testing.T) {
	result := normalizeDigit('5')
	if result != '5' {
		t.Fatalf("expected '5', got '%c'", result)
	}
}

func TestNormalizeDigit_Letter(t *testing.T) {
	result := normalizeDigit('A')
	if result != 'A' {
		t.Fatalf("expected 'A', got '%c'", result)
	}
}

// --- mrzCharValue Tests ---

func TestMRZCharValue_Digit(t *testing.T) {
	result := mrzCharValue('3')
	if result != 3 {
		t.Fatalf("expected 3, got %d", result)
	}
}

func TestMRZCharValue_Letter(t *testing.T) {
	result := mrzCharValue('A')
	if result != 10 {
		t.Fatalf("expected 10, got %d", result)
	}
}

func TestMRZCharValue_Filler(t *testing.T) {
	result := mrzCharValue('<')
	if result != 0 {
		t.Fatalf("expected 0, got %d", result)
	}
}

func TestMRZCharValue_OCorrected(t *testing.T) {
	result := mrzCharValue('O')
	if result != 0 {
		t.Fatalf("expected 0 (O -> 0), got %d", result)
	}
}
