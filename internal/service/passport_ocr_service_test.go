package service

import (
	"testing"
	"time"
)

func TestExtractPlaceOfBirthFromOCRText_SameLineFallback(t *testing.T) {
	dateOfBirth := time.Date(1993, time.July, 26, 0, 0, 0, 0, time.UTC)
	text := "DATE OF BIRTH 26 JUL 93 FENTK SEX F\nNATIONALITY ETH"

	got := extractPlaceOfBirthFromOCRText(text, dateOfBirth)

	if got != "FENTK" {
		t.Fatalf("expected place of birth fallback to return FENTK, got %q", got)
	}
}

func TestExtractPlaceOfBirthFromOCRText_NextLineFallback(t *testing.T) {
	dateOfBirth := time.Date(1993, time.July, 26, 0, 0, 0, 0, time.UTC)
	text := "DATE OF BIRTH 26 JUL 93\nLEGAMBO\nSEX F"

	got := extractPlaceOfBirthFromOCRText(text, dateOfBirth)

	if got != "LEGAMBO" {
		t.Fatalf("expected place of birth fallback to return LEGAMBO, got %q", got)
	}
}

func TestCleanupPlaceOfBirthValue_StripsNoise(t *testing.T) {
	got := cleanupPlaceOfBirthValue("PLACE OF BIRTH FENTK SEX F")

	if got != "FENTK" {
		t.Fatalf("expected cleaned place of birth to be FENTK, got %q", got)
	}
}
