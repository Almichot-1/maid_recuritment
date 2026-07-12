package ocr

// Standard ICAO TD3 test passport (MRP) — valid golden data
const (
	ValidLine1 = "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<"
	ValidLine2 = "L898902C<3UTO6908061F9406236ZE184226B<<<<<14"
)

// MRZ lines with common OCR confusions
const (
	// O instead of 0 in passport number (position 4)
	ConfusedLine2_OFor0 = "L8989O2C<3UTO6908061F9406236ZE184226B<<<<<14"
	// B instead of 8 in passport number (position 3)
	ConfusedLine2_BFor8 = "L89B902C<3UTO6908061F9406236ZE184226B<<<<<14"
	// I instead of 1 in date of birth (position 14 of line2)
	ConfusedLine2_IFor1 = "L898902C<3UTO690806IF9406236ZE184226B<<<<<14"
	// S instead of 5 in date of expiry (position 24 of line2)
	ConfusedLine2_SFor5 = "L898902C<3UTO6908061F94062S6ZE184226B<<<<<14"
)

// Invalid MRZ lines for error-path tests
const (
	ShortLine1         = "SHORT"
	ShortLine2         = "ALSO_SHORT"
	InvalidCharsLine1  = "P<UTØERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<"
	EmptyLine          = ""
)

// Additional valid test passports
const (
	GermanLine1    = "P<D<<STREIBEL<<ALEXANDER<<<<<<<<<<<<<<<<<<<<<"
	GermanLine2    = "C23X00T340DEU7408122M2001023STREIBEL<<<<<<46"
	SingaporeLine1 = "P<SGPLEE<<JOHN<JAMES<<<<<<<<<<<<<<<<<<<<<<<<<<"
	SingaporeLine2 = "E1234567<4SGP8501019M1501010<<<<<<<<<<<<<<08"
)

// ValidGoldenData holds the expected parsed result for the standard test passport.
var ValidGoldenData = &MRZData{
	DocumentType:   "P",
	IssuingCountry: "UTO",
	Surname:        "ERIKSSON",
	GivenNames:     []string{"ANNA", "MARIA"},
	PassportNumber: "L898902C<",
	Nationality:    "UTO",
	Sex:            "F",
	IsValid:        true,
}
