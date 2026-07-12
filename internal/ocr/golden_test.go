package ocr

import "testing"

func TestGolden_ValidPassport(t *testing.T) {
	gc := NewGoldenCase("valid passport", testLine1, testLine2)
	gc.ExpectedDocType = "P"
	gc.ExpectedCountry = "UTO"
	gc.ExpectedSurname = "ERIKSSON"
	gc.ExpectedGivenNames = []string{"ANNA", "MARIA"}
	gc.ExpectedPassport = "L898902C<"
	gc.ExpectedNationality = "UTO"
	gc.ExpectedSex = "F"
	gc.ExpectedIsValid = true
	gc.ExpectedConfidence = 1.0

	if msg := RunGoldenTest(gc); msg != "" {
		t.Fatal(msg)
	}
}

func TestGolden_PassportNumberOCRError(t *testing.T) {
	line2 := "L8989O2C<3UTO6908061F9406236ZE184226B<<<<<14"
	gc := NewGoldenCase("passport number OCR error", testLine1, line2)
	gc.ExpectedPassport = "L898902C<"
	gc.ExpectedIsValid = true
	gc.ExpectedConfidence = 1.0
	gc.ExpectedCorrections = 0

	if msg := RunGoldenTest(gc); msg != "" {
		t.Fatal(msg)
	}
}

func TestGolden_MultipleGivenNames(t *testing.T) {
	line1 := "P<UTOMUSTERMANN<<ERIKA<MARIA<<<<<<<<<<<<<<<<"
	line2 := "C11X00T34<UTO7408122F2001025MUSTERMANN<<<<<<"
	gc := NewGoldenCase("multiple given names", line1, line2)
	gc.ExpectedSurname = "MUSTERMANN"
	gc.ExpectedGivenNames = []string{"ERIKA", "MARIA"}
	gc.ExpectedIsValid = false

	if msg := RunGoldenTest(gc); msg != "" {
		t.Fatal(msg)
	}
}

func TestGolden_InvalidMRZ(t *testing.T) {
	gc := NewGoldenCase("short line1", "SHORT", testLine2)
	gc.ExpectedError = true
	if msg := RunGoldenTest(gc); msg != "" {
		t.Fatal(msg)
	}
}

func TestGolden_EmptyLine1(t *testing.T) {
	gc := NewGoldenCase("empty line1", "", testLine2)
	gc.ExpectedError = true
	if msg := RunGoldenTest(gc); msg != "" {
		t.Fatal(msg)
	}
}

func TestGolden_InvalidCharacters(t *testing.T) {
	line1 := "P<UTØERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<"
	gc := NewGoldenCase("invalid characters", line1, testLine2)
	gc.ExpectedError = true
	if msg := RunGoldenTest(gc); msg != "" {
		t.Fatal(msg)
	}
}
