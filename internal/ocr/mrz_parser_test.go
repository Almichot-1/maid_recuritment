package ocr

import (
	"testing"
	"time"
)

func TestParseMRZBirthDateSelectsPastCentury(t *testing.T) {
	asOf := time.Date(2026, time.March, 28, 0, 0, 0, 0, time.UTC)

	value, err := ParseMRZBirthDate("990213", asOf)
	if err != nil {
		t.Fatalf("ParseMRZBirthDate returned error: %v", err)
	}

	expected := time.Date(1999, time.February, 13, 0, 0, 0, 0, time.UTC)
	if !value.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected.Format(time.DateOnly), value.Format(time.DateOnly))
	}
}

func TestParseMRZExpiryDateSelectsNearestPlausibleCentury(t *testing.T) {
	asOf := time.Date(2026, time.March, 28, 0, 0, 0, 0, time.UTC)

	value, err := ParseMRZExpiryDate("290325", asOf)
	if err != nil {
		t.Fatalf("ParseMRZExpiryDate returned error: %v", err)
	}

	expected := time.Date(2029, time.March, 25, 0, 0, 0, 0, time.UTC)
	if !value.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected.Format(time.DateOnly), value.Format(time.DateOnly))
	}
}

