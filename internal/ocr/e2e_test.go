package ocr

import (
	"strings"
	"testing"
	"time"
)

func buildPassResult(line1, line2 string, conf float64) *MRZPassResult {
	parsed, err := NewMRZParser().ParseMRZ(line1, line2)
	if err != nil {
		return &MRZPassResult{Err: err}
	}
	return &MRZPassResult{
		Parsed:     parsed,
		Confidence: conf,
		Line1:      line1,
		Line2:      line2,
	}
}

func TestPipeline_ParseMRZ_ValidPassport(t *testing.T) {
	p := NewMRZParser()
	data, err := p.ParseMRZ(testLine1, testLine2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.DocumentType != "P" {
		t.Errorf("DocumentType: got %q, want P", data.DocumentType)
	}
	if data.IssuingCountry != "UTO" {
		t.Errorf("IssuingCountry: got %q, want UTO", data.IssuingCountry)
	}
	if data.Surname != "ERIKSSON" {
		t.Errorf("Surname: got %q, want ERIKSSON", data.Surname)
	}
	if len(data.GivenNames) != 2 || data.GivenNames[0] != "ANNA" || data.GivenNames[1] != "MARIA" {
		t.Errorf("GivenNames: got %v, want [ANNA MARIA]", data.GivenNames)
	}
	if data.PassportNumber != "L898902C<" {
		t.Errorf("PassportNumber: got %q, want L898902C<", data.PassportNumber)
	}
	if data.Sex != "F" {
		t.Errorf("Sex: got %q, want F", data.Sex)
	}
	if !data.IsValid {
		t.Errorf("IsValid: got false, want true")
	}
	if data.Confidence <= 0 {
		t.Errorf("Confidence: got %f, want > 0", data.Confidence)
	}
}

func TestPipeline_CheckDigitCorrection(t *testing.T) {
	confusedLine2 := "L89B902C<3UTO6908061F9406236ZE184226B<<<<<14"
	p := NewMRZParser()
	data, err := p.ParseMRZ(testLine1, confusedLine2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.PassportNumber != "L898902C<" {
		t.Errorf("PassportNumber: got %q, want L898902C<", data.PassportNumber)
	}
	if data.CorrectionCount < 1 {
		t.Errorf("CorrectionCount: got %d, want >= 1", data.CorrectionCount)
	}
}

func TestPipeline_FuseMRZResults_Basic(t *testing.T) {
	result1 := buildPassResult(testLine1, testLine2, 0.95)
	confusedLine2 := "L89B902C<3UTO6908061F9406236ZE184226B<<<<<14"
	result2 := buildPassResult(testLine1, confusedLine2, 0.80)

	results := []*MRZPassResult{result1, result2}
	fused := fuseMRZResults(results)

	if fused.DocumentType != "P" {
		t.Errorf("fused DocumentType: got %q, want P", fused.DocumentType)
	}
	if fused.Surname != "ERIKSSON" {
		t.Errorf("fused Surname: got %q, want ERIKSSON", fused.Surname)
	}
	if fused.SuccessfulPasses != 2 {
		t.Errorf("SuccessfulPasses: got %d, want 2", fused.SuccessfulPasses)
	}
	if fused.Confidence <= 0 {
		t.Errorf("fused Confidence: got %f, want > 0", fused.Confidence)
	}
}

func TestPipeline_PostCorrectPassportData(t *testing.T) {
	data := &PassportData{
		Surname:     "SM1TH",
		GivenNames:  "J0HN",
		CountryCode: "US8",
		Confidence:  1.0,
	}
	result := postCorrectPassportData(data)
	if result.Surname != "SMITH" {
		t.Errorf("Surname: got %q, want SMITH", result.Surname)
	}
	if result.GivenNames != "JOHN" {
		t.Errorf("GivenNames: got %q, want JOHN", result.GivenNames)
	}
	if result.Confidence >= 1.0 {
		t.Errorf("Confidence: got %f, want < 1.0", result.Confidence)
	}
}

func TestPipeline_BlurDetection(t *testing.T) {
	sharp := createGrayImage(100, 100, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	score := BlurScore(sharp)
	if score < 100 {
		t.Errorf("sharp image BlurScore: got %f, want >= 100", score)
	}

	uniform := createGrayImage(100, 100, func(x, y int) uint8 {
		return 128
	})
	score = BlurScore(uniform)
	if score > 1 {
		t.Errorf("uniform image BlurScore: got %f, want < 1", score)
	}
}

func TestPipeline_PreprocessMRZ(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		if (x+y)%2 == 0 {
			return 0
		}
		return 255
	})
	result := preprocessPipeline(img, true)
	if result == nil {
		t.Fatal("preprocessPipeline returned nil")
	}
	if result.Bounds().Dx() > 100 || result.Bounds().Dy() > 100 {
		t.Errorf("result dimensions (%dx%d) exceed input (100x100)", result.Bounds().Dx(), result.Bounds().Dy())
	}
}

func TestPipeline_DateValidation(t *testing.T) {
	data := &PassportData{
		DateOfBirth:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		DateOfIssue:  time.Date(1985, 1, 1, 0, 0, 0, 0, time.UTC),
		DateOfExpiry: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	warnings := validatePassportDates(data)
	found := false
	for _, w := range warnings {
		if w == "date of issue must be after date of birth" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning 'date of issue must be after date of birth', got %v", warnings)
	}
}

func TestPipeline_NameCorrection(t *testing.T) {
	data := &PassportData{
		Surname:    "M8R1A",
		Confidence: 1.0,
	}
	result := postCorrectPassportData(data)
	if result.Surname == "M8R1A" {
		t.Errorf("Surname was not corrected, still %q", result.Surname)
	}
	if result.Surname == "" {
		t.Errorf("Surname is empty after correction")
	}
}

func TestPipeline_AccuracyTracking(t *testing.T) {
	expected := &MRZData{
		DocumentType:   "P",
		IssuingCountry: "UTO",
		Surname:        "ERIKSSON",
		GivenNames:     []string{"ANNA", "MARIA"},
		PassportNumber: "L898902C<",
		Nationality:    "UTO",
		Sex:            "F",
		DateOfBirth:    time.Date(1969, 8, 6, 0, 0, 0, 0, time.UTC),
		DateOfExpiry:   time.Date(1994, 6, 23, 0, 0, 0, 0, time.UTC),
	}
	actual := &MRZData{
		DocumentType:   "P",
		IssuingCountry: "UTO",
		Surname:        "ERIKSSON",
		GivenNames:     []string{"ANNA", "MARIA"},
		PassportNumber: "L898902C<",
		Nationality:    "UTO",
		Sex:            "F",
		DateOfBirth:    time.Date(1969, 8, 6, 0, 0, 0, 0, time.UTC),
		DateOfExpiry:   time.Date(1994, 6, 23, 0, 0, 0, 0, time.UTC),
	}
	results := CompareFields(actual, expected)
	if results == nil {
		t.Fatal("CompareFields returned nil")
	}
	for field, correct := range results {
		if !correct {
			t.Errorf("field %q: got incorrect, want correct", field)
		}
	}

	incorrectActual := &MRZData{
		DocumentType:   "V",
		IssuingCountry: "XYZ",
		Surname:        "WRONG",
		GivenNames:     []string{"WRONG"},
		PassportNumber: "XXXXXXXXX",
		Nationality:    "XYZ",
		Sex:            "M",
		DateOfBirth:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		DateOfExpiry:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	results2 := CompareFields(incorrectActual, expected)
	correctCount := 0
	for _, correct := range results2 {
		if correct {
			correctCount++
		}
	}
	if correctCount > 0 {
		t.Errorf("expected 0 correct fields for fully mismatched data, got %d", correctCount)
	}
}

func TestPipeline_EndToEnd(t *testing.T) {
	p := NewMRZParser()
	parsed, err := p.ParseMRZ(testLine1, testLine2)
	if err != nil {
		t.Fatalf("ParseMRZ failed: %v", err)
	}

	result := &MRZPassResult{
		Parsed:     parsed,
		Confidence: parsed.Confidence,
		Line1:      testLine1,
		Line2:      testLine2,
	}
	fused := fuseMRZResults([]*MRZPassResult{result})
	if fused.Confidence <= 0 {
		t.Errorf("fused Confidence: got %f, want > 0", fused.Confidence)
	}

	passportData := &PassportData{
		DocumentType:   fused.DocumentType,
		CountryCode:    fused.IssuingCountry,
		Surname:        fused.Surname,
		GivenNames:     strings.Join(fused.GivenNames, " "),
		PassportNumber: fused.PassportNumber,
		Nationality:    fused.Nationality,
		DateOfBirth:    fused.DateOfBirth,
		Sex:            fused.Sex,
		DateOfExpiry:   fused.DateOfExpiry,
		MRZLine1:       fused.Line1,
		MRZLine2:       fused.Line2,
		Confidence:     fused.Confidence,
		ExtractedAt:    time.Now().UTC(),
	}

	_ = postCorrectPassportData(passportData)

	if passportData.Confidence <= 0 {
		t.Errorf("final Confidence: got %f, want > 0", passportData.Confidence)
	}
	if passportData.Surname == "" {
		t.Errorf("Surname is empty")
	}
	if passportData.GivenNames == "" {
		t.Errorf("GivenNames is empty")
	}
	if passportData.DocumentType == "" {
		t.Errorf("DocumentType is empty")
	}
	if passportData.CountryCode == "" {
		t.Errorf("CountryCode is empty")
	}
}
