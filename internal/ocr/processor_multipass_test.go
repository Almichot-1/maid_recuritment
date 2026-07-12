package ocr

import (
	"os"
	"testing"
	"time"
)

func makePassResult(conf float64, err error) *MRZPassResult {
	data := &MRZData{
		Confidence: conf,
		IsValid:    true,
		RawLine1:   testLine1,
		RawLine2:   testLine2,
	}
	return &MRZPassResult{
		Parsed:     data,
		Confidence: conf,
		Err:        err,
	}
}

func makeResult(passport string, dob time.Time, conf float64, line1, line2 string) *MRZPassResult {
	data := &MRZData{
		DocumentType:   "P",
		IssuingCountry: "UTO",
		Surname:        "ERIKSSON",
		GivenNames:     []string{"ANNA", "MARIA"},
		PassportNumber: passport,
		Nationality:    "UTO",
		DateOfBirth:    dob,
		DateOfExpiry:   time.Date(2029, time.March, 25, 0, 0, 0, 0, time.UTC),
		Sex:            "F",
		OptionalData:   "",
		RawLine1:       line1,
		RawLine2:       line2,
		Confidence:     conf,
		IsValid:        true,
	}
	return &MRZPassResult{
		Parsed:     data,
		Confidence: conf,
		Line1:      line1,
		Line2:      line2,
	}
}

func TestBestSinglePassResult_NilSlice(t *testing.T) {
	result := bestSinglePassResult(nil)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestBestSinglePassResult_AllErrors(t *testing.T) {
	results := []*MRZPassResult{
		makePassResult(0.5, os.ErrInvalid),
		makePassResult(0.7, os.ErrInvalid),
		makePassResult(0.9, os.ErrInvalid),
	}
	result := bestSinglePassResult(results)
	if result != nil {
		t.Fatalf("expected nil for all errors, got confidence %f", result.Confidence)
	}
}

func TestBestSinglePassResult_SelectsHighestConfidence(t *testing.T) {
	results := []*MRZPassResult{
		makePassResult(0.3, nil),
		makePassResult(0.7, nil),
		makePassResult(0.9, nil),
	}
	result := bestSinglePassResult(results)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Confidence != 0.9 {
		t.Fatalf("expected confidence 0.9, got %f", result.Confidence)
	}
}

func TestBestSinglePassResult_SkipsErrors(t *testing.T) {
	results := []*MRZPassResult{
		makePassResult(0.9, os.ErrInvalid),
		makePassResult(0.5, nil),
	}
	result := bestSinglePassResult(results)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Confidence != 0.5 {
		t.Fatalf("expected confidence 0.5, got %f", result.Confidence)
	}
}

func TestValidPassResults_FiltersErrors(t *testing.T) {
	results := []*MRZPassResult{
		makePassResult(0.9, os.ErrInvalid),
		makePassResult(0.7, nil),
		makePassResult(0.5, os.ErrInvalid),
	}
	valid := validPassResults(results)
	if len(valid) != 1 {
		t.Fatalf("expected 1 valid result, got %d", len(valid))
	}
	if valid[0].Confidence != 0.7 {
		t.Fatalf("expected confidence 0.7, got %f", valid[0].Confidence)
	}
}

func TestValidPassResults_AllNil(t *testing.T) {
	results := []*MRZPassResult{
		{Err: nil, Parsed: nil, Confidence: 0.9},
		{Err: nil, Parsed: nil, Confidence: 0.7},
	}
	valid := validPassResults(results)
	if len(valid) != 0 {
		t.Fatalf("expected 0 valid results, got %d", len(valid))
	}
}

func TestValidPassResults_AllSuccess(t *testing.T) {
	results := []*MRZPassResult{
		makePassResult(0.9, nil),
		makePassResult(0.7, nil),
		makePassResult(0.5, nil),
	}
	valid := validPassResults(results)
	if len(valid) != 3 {
		t.Fatalf("expected 3 valid results, got %d", len(valid))
	}
}

func TestFuseMRZResults_NilInput(t *testing.T) {
	fused := fuseMRZResults(nil)
	if fused == nil {
		t.Fatal("expected non-nil FusedMRZResult")
	}
	if fused.TotalPasses != 0 {
		t.Fatalf("expected TotalPasses 0, got %d", fused.TotalPasses)
	}
	if fused.SuccessfulPasses != 0 {
		t.Fatalf("expected SuccessfulPasses 0, got %d", fused.SuccessfulPasses)
	}
}

func TestFuseMRZResults_SinglePass(t *testing.T) {
	dob := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	result := makeResult("L898902C<", dob, 0.9, testLine1, testLine2)
	fused := fuseMRZResults([]*MRZPassResult{result})

	if fused.TotalPasses != 1 {
		t.Fatalf("expected TotalPasses 1, got %d", fused.TotalPasses)
	}
	if fused.SuccessfulPasses != 1 {
		t.Fatalf("expected SuccessfulPasses 1, got %d", fused.SuccessfulPasses)
	}
	if fused.PassportNumber != "L898902C<" {
		t.Fatalf("expected passport L898902C<, got %q", fused.PassportNumber)
	}
	if fused.DocumentType != "P" {
		t.Fatalf("expected DocumentType P, got %q", fused.DocumentType)
	}
	if fused.IssuingCountry != "UTO" {
		t.Fatalf("expected IssuingCountry UTO, got %q", fused.IssuingCountry)
	}
	if fused.Surname != "ERIKSSON" {
		t.Fatalf("expected Surname ERIKSSON, got %q", fused.Surname)
	}
	if fused.Nationality != "UTO" {
		t.Fatalf("expected Nationality UTO, got %q", fused.Nationality)
	}
	if fused.Sex != "F" {
		t.Fatalf("expected Sex F, got %q", fused.Sex)
	}
	if len(fused.GivenNames) != 2 || fused.GivenNames[0] != "ANNA" || fused.GivenNames[1] != "MARIA" {
		t.Fatalf("expected GivenNames [ANNA MARIA], got %v", fused.GivenNames)
	}
	if !fused.DateOfBirth.Equal(dob) {
		t.Fatalf("expected DOB %s, got %s", dob.Format(time.DateOnly), fused.DateOfBirth.Format(time.DateOnly))
	}
	if fused.Line1 != testLine1 {
		t.Fatalf("expected Line1 %q, got %q", testLine1, fused.Line1)
	}
	if fused.Line2 != testLine2 {
		t.Fatalf("expected Line2 %q, got %q", testLine2, fused.Line2)
	}
}

func TestFuseMRZResults_ChoosesBestPerField(t *testing.T) {
	dob := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	passA := makeResult("L898902C<", dob, 0.9, testLine1, testLine2)
	passB := makeResult("L898902C<", dob, 0.7, testLine1+"X", testLine2+"X")

	fused := fuseMRZResults([]*MRZPassResult{passA, passB})

	if fused.PassportNumber != "L898902C<" {
		t.Fatalf("expected passport L898902C<, got %q", fused.PassportNumber)
	}
	if !fused.DateOfBirth.Equal(dob) {
		t.Fatalf("expected DOB %s, got %s", dob.Format(time.DateOnly), fused.DateOfBirth.Format(time.DateOnly))
	}
	if fused.Line1 != testLine1 {
		t.Fatalf("expected Line1 from highest-confidence pass, got %q", fused.Line1)
	}
	if fused.Line2 != testLine2 {
		t.Fatalf("expected Line2 from highest-confidence pass, got %q", fused.Line2)
	}
}

func TestFuseMRZResults_MissingFields(t *testing.T) {
	dob := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	passA := makeResult("", dob, 0.9, testLine1, testLine2)
	passB := makeResult("L898902C<", dob, 0.7, testLine1, testLine2)

	fused := fuseMRZResults([]*MRZPassResult{passA, passB})

	if fused.PassportNumber != "L898902C<" {
		t.Fatalf("expected passport L898902C< (fallback from pass B), got %q", fused.PassportNumber)
	}
	if !fused.DateOfBirth.Equal(dob) {
		t.Fatalf("expected DOB %s, got %s", dob.Format(time.DateOnly), fused.DateOfBirth.Format(time.DateOnly))
	}
	if fused.Line1 != testLine1 {
		t.Fatalf("expected Line1 from highest-confidence pass, got %q", fused.Line1)
	}
	if fused.Line2 != testLine2 {
		t.Fatalf("expected Line2 from highest-confidence pass, got %q", fused.Line2)
	}
}

func TestFuseMRZResults_LineFromBest(t *testing.T) {
	dob := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	line1A, line2A := "LINE1_FROM_PASS_A", "LINE2_FROM_PASS_A"
	line1B, line2B := "LINE1_FROM_PASS_B", "LINE2_FROM_PASS_B"
	passA := makeResult("L898902C<", dob, 0.9, line1A, line2A)
	passB := makeResult("L898902C<", dob, 0.7, line1B, line2B)

	fused := fuseMRZResults([]*MRZPassResult{passA, passB})

	if fused.Line1 != line1A {
		t.Fatalf("expected Line1 %q from highest-confidence pass, got %q", line1A, fused.Line1)
	}
	if fused.Line2 != line2A {
		t.Fatalf("expected Line2 %q from highest-confidence pass, got %q", line2A, fused.Line2)
	}
}

func TestFuseMRZResults_FieldConfidence(t *testing.T) {
	dob := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	passA := makeResult("ABC123", dob, 0.9, testLine1, testLine2)
	passB := makeResult("ABC123", dob, 0.7, testLine1, testLine2)

	passA.Parsed.Surname = "SMITH"
	passB.Parsed.Surname = "SMITH"

	fused := fuseMRZResults([]*MRZPassResult{passA, passB})

	if fused.FieldConfidence["passport_number"] != 1.0 {
		t.Fatalf("expected passport_number field confidence 1.0 (both agree), got %f", fused.FieldConfidence["passport_number"])
	}
	if fused.FieldConfidence["document_type"] != 1.0 {
		t.Fatalf("expected document_type field confidence 1.0, got %f", fused.FieldConfidence["document_type"])
	}
	if fused.FieldConfidence["surname"] != 1.0 {
		t.Fatalf("expected surname field confidence 1.0 (both agree), got %f", fused.FieldConfidence["surname"])
	}
	if fused.FieldConfidence["date_of_birth"] != 1.0 {
		t.Fatalf("expected date_of_birth field confidence 1.0 (both agree), got %f", fused.FieldConfidence["date_of_birth"])
	}
}

func TestFuseMRZResults_FieldConfidenceMixed(t *testing.T) {
	dob := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	passA := makeResult("ABC123", dob, 0.9, testLine1, testLine2)
	passB := makeResult("XYZ789", dob, 0.7, testLine1, testLine2)

	fused := fuseMRZResults([]*MRZPassResult{passA, passB})

	if fused.FieldConfidence["passport_number"] != 0.5 {
		t.Fatalf("expected passport_number field confidence 0.5 (split vote), got %f", fused.FieldConfidence["passport_number"])
	}
	if fused.FieldConfidence["date_of_birth"] != 1.0 {
		t.Fatalf("expected date_of_birth field confidence 1.0 (both agree), got %f", fused.FieldConfidence["date_of_birth"])
	}
	if fused.FieldConfidence["document_type"] != 1.0 {
		t.Fatalf("expected document_type field confidence 1.0, got %f", fused.FieldConfidence["document_type"])
	}
}

func TestFindLowConfidenceFields_AllAbove(t *testing.T) {
	fc := map[string]float64{
		"passport_number": 0.9,
		"date_of_birth":   0.85,
		"surname":         1.0,
	}
	low := findLowConfidenceFields(fc)
	if len(low) != 0 {
		t.Fatalf("expected 0 low-confidence fields, got %d: %v", len(low), low)
	}
}

func TestFindLowConfidenceFields_SomeBelow(t *testing.T) {
	fc := map[string]float64{
		"passport_number": 0.9,
		"date_of_birth":   0.7,
		"surname":         0.5,
	}
	low := findLowConfidenceFields(fc)
	if len(low) != 2 {
		t.Fatalf("expected 2 low-confidence fields, got %d: %v", len(low), low)
	}
	found := make(map[string]bool)
	for _, f := range low {
		found[f] = true
	}
	if !found["date_of_birth"] {
		t.Fatal("expected date_of_birth in low-confidence fields")
	}
	if !found["surname"] {
		t.Fatal("expected surname in low-confidence fields")
	}
}

func TestFindLowConfidenceFields_AllBelow(t *testing.T) {
	fc := map[string]float64{
		"passport_number": 0.3,
		"date_of_birth":   0.5,
		"surname":         0.1,
	}
	low := findLowConfidenceFields(fc)
	if len(low) != 3 {
		t.Fatalf("expected 3 low-confidence fields, got %d: %v", len(low), low)
	}
}

func TestFindLowConfidenceFields_NilMap(t *testing.T) {
	low := findLowConfidenceFields(nil)
	if low != nil {
		t.Fatalf("expected nil for nil map, got %v", low)
	}
}

func TestDefaultMRZPassConfigs_NotEmpty(t *testing.T) {
	configs := defaultMRZPassConfigs()
	if len(configs) < 3 {
		t.Fatalf("expected at least 3 configs, got %d", len(configs))
	}
}

func TestDefaultMRZPassConfigs_HasPSM6OEM1(t *testing.T) {
	configs := defaultMRZPassConfigs()
	found := false
	for _, c := range configs {
		if c.PSMMode == 6 && c.OEMMode == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected config with PSMMode=6 and OEMMode=1")
	}
}

func TestExtractMRZ_EmptyPath(t *testing.T) {
	p := NewOCRProcessor("", "eng")
	_, _, _, err := p.ExtractMRZ("")
	if err == nil {
		t.Fatal("expected error for empty image path")
	}
}
