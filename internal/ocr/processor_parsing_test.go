package ocr

import (
	"testing"
	"time"
)

// --- Levenshtein Distance Tests ---

func TestLevenshteinDistance_Identical(t *testing.T) {
	dist := levenshteinDistance("PLACE OF BIRTH", "PLACE OF BIRTH")
	if dist != 0 {
		t.Fatalf("expected 0 for identical strings, got %d", dist)
	}
}

func TestLevenshteinDistance_OneSubstitution(t *testing.T) {
	dist := levenshteinDistance("PLACE OF BIRTH", "PLACE OF B1RTH")
	if dist != 1 {
		t.Fatalf("expected 1 for one substitution, got %d", dist)
	}
}

func TestLevenshteinDistance_TwoSubstitutions(t *testing.T) {
	dist := levenshteinDistance("PLACE OF BIRTH", "PLACE OF B1R7H")
	if dist != 2 {
		t.Fatalf("expected 2 for two substitutions, got %d", dist)
	}
}

func TestLevenshteinDistance_OneInsertion(t *testing.T) {
	dist := levenshteinDistance("BIRTH", "BIRTHS")
	if dist != 1 {
		t.Fatalf("expected 1 for one insertion, got %d", dist)
	}
}

func TestLevenshteinDistance_OneDeletion(t *testing.T) {
	dist := levenshteinDistance("BIRTHS", "BIRTH")
	if dist != 1 {
		t.Fatalf("expected 1 for one deletion, got %d", dist)
	}
}

func TestLevenshteinDistance_EmptyFirst(t *testing.T) {
	dist := levenshteinDistance("", "BIRTH")
	if dist != 5 {
		t.Fatalf("expected 5 for empty to BIRTH, got %d", dist)
	}
}

func TestLevenshteinDistance_EmptySecond(t *testing.T) {
	dist := levenshteinDistance("BIRTH", "")
	if dist != 5 {
		t.Fatalf("expected 5 for BIRTH to empty, got %d", dist)
	}
}

func TestLevenshteinDistance_BothEmpty(t *testing.T) {
	dist := levenshteinDistance("", "")
	if dist != 0 {
		t.Fatalf("expected 0 for both empty, got %d", dist)
	}
}

// --- extractVisualZoneFields Tests ---

func TestExtractVisualZoneFields_PlaceOfBirth(t *testing.T) {
	text := "PLACE OF BIRTH ADDIS ABABA\nSEX F\nNATIONALITY ETHIOPIAN"
	vz := extractVisualZoneFields(text, 0)
	if vz.PlaceOfBirth == "" {
		t.Fatal("expected place of birth to be extracted")
	}
}

func TestExtractVisualZoneFields_PlaceOfBirthWithOCRError(t *testing.T) {
	// Fuzzy match (Levenshtein) handles 1→I substitution.
	// The value is on the next line after the fuzzy-matched label.
	text := "PLACE OF B1RTH\nADDIS ABABA\nSEX F\nNATIONALITY ETHIOPIAN"
	vz := extractVisualZoneFields(text, 0)
	if vz.PlaceOfBirth == "" {
		t.Fatal("expected place of birth to be extracted despite OCR error")
	}
}

func TestExtractVisualZoneFields_IssuingAuthority(t *testing.T) {
	text := "ISSUING AUTHORITY DEPARTMENT OF IMMIGRATION\nDATE OF ISSUE 15 JAN 2020"
	vz := extractVisualZoneFields(text, 0)
	if vz.Authority == "" {
		t.Fatal("expected issuing authority to be extracted")
	}
}

func TestExtractVisualZoneFields_DateOfIssue(t *testing.T) {
	text := "DATE OF ISSUE 15 JAN 2020\nDATE OF BIRTH 01 JAN 1990"
	vz := extractVisualZoneFields(text, 0)
	if vz.DateOfIssue.IsZero() {
		t.Fatal("expected date of issue to be extracted")
	}
}

func TestExtractVisualZoneFields_InternationalLabel(t *testing.T) {
	text := "LIEU DE NAISSANCE PARIS\nSEXE F\nNATIONALITE FRANCAISE"
	vz := extractVisualZoneFields(text, 0)
	if vz.PlaceOfBirth == "" {
		t.Fatal("expected place of birth to be extracted from French label")
	}
}

func TestExtractVisualZoneFields_NoMatch(t *testing.T) {
	text := "SOME RANDOM TEXT\nMORE TEXT\nNO STRUCTURED FIELDS HERE"
	vz := extractVisualZoneFields(text, 0)
	if vz.PlaceOfBirth != "" || !vz.DateOfIssue.IsZero() {
		t.Fatal("expected no fields to be extracted from garbage text")
	}
}

// --- findLabelValue Fuzzy Matching Tests ---

func TestFindLabelValueViaExtractVisualZone_FuzzyMatch(t *testing.T) {
	text := "ISSUING AUTH0R1TY MINISTRY OF INTERNAL AFFAIRS\nDATE OF ISSUE 15 JAN 2020"
	vz := extractVisualZoneFields(text, 0)
	if vz.Authority == "" {
		t.Fatal("expected issuing authority to be extracted via fuzzy match")
	}
}

// --- findDateOfIssue Tests ---

func TestFindDateOfIssue_ISODate(t *testing.T) {
	text := "DATE OF ISSUE 2020-01-15\nSOME OTHER TEXT"
	vz := extractVisualZoneFields(text, 0)
	expected := time.Date(2020, time.January, 15, 0, 0, 0, 0, time.UTC)
	if !vz.DateOfIssue.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected.Format(time.DateOnly), vz.DateOfIssue.Format(time.DateOnly))
	}
}

func TestFindDateOfIssue_SlashDate(t *testing.T) {
	text := "DATE OF ISSUE 15/01/2020\nSOME OTHER TEXT"
	vz := extractVisualZoneFields(text, 0)
	expected := time.Date(2020, time.January, 15, 0, 0, 0, 0, time.UTC)
	if !vz.DateOfIssue.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected.Format(time.DateOnly), vz.DateOfIssue.Format(time.DateOnly))
	}
}

func TestFindDateOfIssue_InternationalLabel(t *testing.T) {
	text := "DATE D'EMISSION 15/01/2020\nLIEU DE NAISSANCE PARIS"
	vz := extractVisualZoneFields(text, 0)
	expected := time.Date(2020, time.January, 15, 0, 0, 0, 0, time.UTC)
	if !vz.DateOfIssue.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected.Format(time.DateOnly), vz.DateOfIssue.Format(time.DateOnly))
	}
}

// --- cleanupVisualValue Tests ---

func TestCleanupVisualValue_SingleWord(t *testing.T) {
	result := cleanupVisualValue("ADDIS ABABA")
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestCleanupVisualValue_WithOCRGarbage(t *testing.T) {
	result := cleanupVisualValue("ADDIS<ABABA")
	if result != "ADDIS ABABA" {
		t.Fatalf("expected 'ADDIS ABABA', got %q", result)
	}
}

func TestCleanupVisualValue_ShortWordsFiltered(t *testing.T) {
	result := cleanupVisualValue("AB CD EF")
	if result != "" {
		t.Fatalf("expected empty (all words too short), got %q", result)
	}
}

// --- containsAny Tests ---

func TestContainsAny_Matches(t *testing.T) {
	if !containsAny("DATE OF ISSUE", []string{"ISSUE DATE", "DATE OF ISSUE"}) {
		t.Fatal("expected match")
	}
}

func TestContainsAny_NoMatch(t *testing.T) {
	if containsAny("SOME TEXT", []string{"DATE", "ISSUE"}) {
		t.Fatal("expected no match")
	}
}

func TestContainsAny_EmptyStrings(t *testing.T) {
	if !containsAny("", []string{""}) {
		t.Fatal("expected match for empty strings (strings.Contains(\"\", \"\") is true)")
	}
}

// --- findPlaceOfBirthNearBirthDate Tests ---

func TestFindPlaceOfBirthNearBirthDate_SameLine(t *testing.T) {
	dob := time.Date(1993, time.July, 26, 0, 0, 0, 0, time.UTC)
	text := "DATE OF BIRTH 26 JUL 93 FENTK SEX F\nNATIONALITY ETH"
	result := findPlaceOfBirthNearBirthDate(text, dob)
	if result == "" {
		t.Fatal("expected place of birth to be found near birth date")
	}
}

func TestFindPlaceOfBirthNearBirthDate_NextLine(t *testing.T) {
	dob := time.Date(1993, time.July, 26, 0, 0, 0, 0, time.UTC)
	text := "DATE OF BIRTH 26 JUL 93\nLEGAMBO\nSEX F"
	result := findPlaceOfBirthNearBirthDate(text, dob)
	if result == "" {
		t.Fatal("expected place of birth on next line")
	}
}
