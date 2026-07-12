package ocr

import (
	"image/color"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Assertion helpers
// ---------------------------------------------------------------------------

func assertEqual(t *testing.T, got, want, field string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", field, got, want)
	}
}

func assertTrue(t *testing.T, val bool, msg string) {
	t.Helper()
	if !val {
		t.Error(msg)
	}
}

func assertNotNil(t *testing.T, val interface{}, msg string) {
	t.Helper()
	if val == nil {
		t.Error(msg)
	}
}

// ---------------------------------------------------------------------------
// Phase 1 — Check digit correction
// ---------------------------------------------------------------------------

func TestRegression_CheckDigitCorrection_BTo8(t *testing.T) {
	field := "L89B902C"
	corrected, count, err := CorrectCheckDigitField(field, '3')
	if err != nil {
		t.Fatalf("expected correction to succeed: %v", err)
	}
	assertEqual(t, corrected, "L898902C", "corrected passport number")
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestRegression_CheckDigitCorrection_OTo0(t *testing.T) {
	cleaned, err := CleanMRZLine(ConfusedLine2_OFor0)
	if err != nil {
		t.Fatalf("CleanMRZLine failed: %v", err)
	}
	// The O at position 4 should become 0 when adjacent to digits
	expectedPassportPart := "L898902C<"
	if len(cleaned) < 9 {
		t.Fatalf("cleaned line too short: %q", cleaned)
	}
	assertEqual(t, cleaned[:9], expectedPassportPart, "passport portion after O→0 fix")
}

func TestRegression_CheckDigitCorrection_AlreadyValid(t *testing.T) {
	_, _, err := CorrectCheckDigitField("L898902C", '3')
	if err == nil {
		t.Fatal("expected error for already-valid field")
	}
}

func TestRegression_ConfidenceScoring(t *testing.T) {
	p := NewMRZParser()
	data, err := p.ParseMRZ(ValidLine1, ValidLine2)
	if err != nil {
		t.Fatalf("ParseMRZ failed: %v", err)
	}
	conf := ComputeConfidence(data)
	if conf != 1.0 {
		t.Fatalf("expected confidence 1.0, got %f", conf)
	}
}

// ---------------------------------------------------------------------------
// Phase 2 — Image preprocessing
// ---------------------------------------------------------------------------

func TestRegression_BlurDetection(t *testing.T) {
	sharp := createGrayImage(100, 100, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	score := BlurScore(sharp)
	if score < 100 {
		t.Fatalf("expected sharp image score >= 100, got %f", score)
	}

	uniform := createGrayImage(100, 100, func(x, y int) uint8 {
		return 128
	})
	score = BlurScore(uniform)
	if score > 1 {
		t.Fatalf("expected uniform image score near 0, got %f", score)
	}
}

func TestRegression_SauvolaThreshold(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	result := sauvolaThresholdDefault(img)
	assertNotNil(t, result, "sauvola result should not be nil")

	hasBlack, hasWhite := false, false
	for _, v := range result.Pix {
		if v == 0 {
			hasBlack = true
		}
		if v == 255 {
			hasWhite = true
		}
	}
	assertTrue(t, hasBlack, "expected black pixels")
	assertTrue(t, hasWhite, "expected white pixels")
}

func TestRegression_CLAHE(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		return uint8(100 + (x+y)%21)
	})
	result := applyCLAHE(img, 3.0, 8)
	assertNotNil(t, result, "CLAHE result should not be nil")

	min, max := uint8(255), uint8(0)
	for _, v := range result.Pix {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if max-min <= 20 {
		t.Fatalf("expected contrast enhancement, got range [%d,%d]", min, max)
	}
}

func TestRegression_MedianDenoise(t *testing.T) {
	img := createGrayImage(50, 50, func(x, y int) uint8 {
		return 255
	})
	img.SetGray(0, 0, color.Gray{Y: 0})
	img.SetGray(1, 1, color.Gray{Y: 0})
	denoised := medianDenoise(img, 3)
	assertNotNil(t, denoised, "denoised image should not be nil")
}

func TestRegression_PreprocessPipeline(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 30
		}
		return 200
	})
	result := preprocessPipeline(img, true)
	assertNotNil(t, result, "pipeline output should not be nil")

	resultNoBinarize := preprocessPipeline(img, false)
	assertNotNil(t, resultNoBinarize, "pipeline output (non-MRZ) should not be nil")
}

// ---------------------------------------------------------------------------
// Phase 3 — Multi-pass fusion
// ---------------------------------------------------------------------------

func TestRegression_FuseMRZResults(t *testing.T) {
	dob := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	passA := makeResult("ABC123", dob, 0.9, ValidLine1, ValidLine2)
	passB := makeResult("ABC123", dob, 0.7, ValidLine1, ValidLine2)

	fused := fuseMRZResults([]*MRZPassResult{passA, passB})
	assertNotNil(t, fused, "fused result should not be nil")
	assertEqual(t, fused.PassportNumber, "ABC123", "fused passport number")
	assertEqual(t, fused.DocumentType, "P", "fused document type")
	assertEqual(t, fused.Surname, "ERIKSSON", "fused surname")

	if fused.Confidence <= 0 {
		t.Fatal("expected positive fused confidence")
	}
}

func TestRegression_FuseMRZResults_NilInput(t *testing.T) {
	fused := fuseMRZResults(nil)
	assertNotNil(t, fused, "fused result should be non-nil even with nil input")
	if fused.TotalPasses != 0 {
		t.Fatalf("expected TotalPasses 0, got %d", fused.TotalPasses)
	}
	if fused.SuccessfulPasses != 0 {
		t.Fatalf("expected SuccessfulPasses 0, got %d", fused.SuccessfulPasses)
	}
}

// ---------------------------------------------------------------------------
// Phase 4 — Post-correction
// ---------------------------------------------------------------------------

func TestRegression_CountryCodeValidation(t *testing.T) {
	assertTrue(t, isValidCountryCode("USA"), "USA should be valid")
	assertTrue(t, isValidCountryCode("GBR"), "GBR should be valid")
	assertTrue(t, isValidCountryCode("SGP"), "SGP should be valid")
	assertTrue(t, !isValidCountryCode("XYZ"), "XYZ should be invalid")
	assertTrue(t, !isValidCountryCode("A12"), "A12 should be invalid")
	assertTrue(t, !isValidCountryCode(""), "empty should be invalid")
}

func TestRegression_NameOCRCrorrection(t *testing.T) {
	result, changed := correctNameOcr("SM1TH")
	assertTrue(t, changed, "expected change for SM1TH")
	assertEqual(t, result, "SMITH", "corrected name")

	_, changed = correctNameOcr("SMITH")
	assertTrue(t, !changed, "expected no change for correct name")
}

func TestRegression_DateValidation(t *testing.T) {
	dob := testDate(1990, 1, 1)
	issue := testDate(2010, 6, 15)
	expiry := testDate(2030, 6, 14)
	warnings := validateDateConsistency(dob, issue, expiry)
	if warnings != nil {
		t.Fatalf("expected nil for valid dates, got %v", warnings)
	}

	dobFarPast := testDate(1850, 1, 1)
	warnings = validateDateConsistency(dobFarPast, issue, expiry)
	if len(warnings) < 1 {
		t.Fatal("expected warnings for unreasonably old birth date")
	}
}

func TestRegression_FullPostCorrection(t *testing.T) {
	data := &PassportData{
		Surname:    "SM1TH",
		GivenNames: "J0HN",
		CountryCode: "XYZ",
		Nationality: "USA",
		Confidence: 1.0,
	}
	result := postCorrectPassportData(data)
	assertNotNil(t, result, "post-correct result should not be nil")
	assertEqual(t, result.Surname, "SMITH", "corrected surname")
	assertEqual(t, result.GivenNames, "JOHN", "corrected given names")
	if result.Confidence >= 1.0 {
		t.Fatal("expected confidence penalty for unresolved invalid country code")
	}
}

// ---------------------------------------------------------------------------
// Comprehensive MRZ parse regression
// ---------------------------------------------------------------------------

func TestRegression_MRZParse_AllFields(t *testing.T) {
	p := NewMRZParser()
	data, err := p.ParseMRZ(ValidLine1, ValidLine2)
	if err != nil {
		t.Fatalf("ParseMRZ failed: %v", err)
	}

	assertEqual(t, data.DocumentType, "P", "DocumentType")
	assertEqual(t, data.IssuingCountry, "UTO", "IssuingCountry")
	assertEqual(t, data.Surname, "ERIKSSON", "Surname")

	if len(data.GivenNames) != 2 {
		t.Fatalf("expected 2 given names, got %d", len(data.GivenNames))
	}
	assertEqual(t, data.GivenNames[0], "ANNA", "GivenName[0]")
	assertEqual(t, data.GivenNames[1], "MARIA", "GivenName[1]")
	assertEqual(t, data.PassportNumber, "L898902C<", "PassportNumber")

	expectedDOB := time.Date(1969, time.August, 6, 0, 0, 0, 0, time.UTC)
	if !data.DateOfBirth.Equal(expectedDOB) {
		t.Fatalf("DateOfBirth: got %s, want %s", data.DateOfBirth.Format(time.DateOnly), expectedDOB.Format(time.DateOnly))
	}

	assertEqual(t, data.Sex, "F", "Sex")

	expectedExpiry := time.Date(1994, time.June, 23, 0, 0, 0, 0, time.UTC)
	if !data.DateOfExpiry.Equal(expectedExpiry) {
		t.Fatalf("DateOfExpiry: got %s, want %s", data.DateOfExpiry.Format(time.DateOnly), expectedExpiry.Format(time.DateOnly))
	}

	assertEqual(t, data.Nationality, "UTO", "Nationality")
	assertTrue(t, data.IsValid, "IsValid should be true")
	assertEqual(t, data.OptionalData, "ZE184226B<<<<<", "OptionalData")
}
