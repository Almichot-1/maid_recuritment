package ocr

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type MRZParser struct{}

func NewMRZParser() *MRZParser {
	return &MRZParser{}
}

type MRZData struct {
	CorrectionCount int      // number of check digit corrections applied
	Confidence      float64  // 0.0-1.0 based on validation ratio
	DocumentType     string
	IssuingCountry   string
	Surname          string
	GivenNames       []string
	PassportNumber   string
	Nationality      string
	DateOfBirth      time.Time
	Sex              string
	DateOfExpiry     time.Time
	OptionalData     string
	RawLine1         string
	RawLine2         string
	IsValid          bool
	ValidationErrors []string
}

func (p *MRZParser) ParseMRZ(line1, line2 string) (*MRZData, error) {
	l1, err := CleanMRZLine(line1)
	if err != nil {
		return nil, err
	}
	l2, err := CleanMRZLine(line2)
	if err != nil {
		return nil, err
	}

	if len(l1) != 44 {
		return nil, fmt.Errorf("MRZ line1 must be 44 characters")
	}
	if len(l2) != 44 {
		return nil, fmt.Errorf("MRZ line2 must be 44 characters")
	}

	data := &MRZData{RawLine1: l1, RawLine2: l2, IsValid: true}
	data.CorrectionCount = 0
	data.DocumentType = l1[0:1]
	data.IssuingCountry = l1[2:5]
	data.Surname, data.GivenNames = ParseName(l1[5:44])

	passportNumber := l2[0:9]
	passportCD := l2[9]
	nationality := l2[10:13]
	dobStr := l2[13:19]
	dobCD := l2[19]
	sex := l2[20:21]
	expStr := l2[21:27]
	expCD := l2[27]
	optional := l2[28:42]
	optionalCD := l2[42]
	compositeCD := l2[43]

	data.PassportNumber = passportNumber
	data.Nationality = nationality
	data.Sex = sex
	data.OptionalData = optional

	if dob, derr := ParseMRZBirthDate(dobStr, time.Now().UTC()); derr != nil {
		data.IsValid = false
		data.ValidationErrors = append(data.ValidationErrors, "invalid date of birth")
	} else {
		data.DateOfBirth = dob
	}
	if exp, eerr := ParseMRZExpiryDate(expStr, time.Now().UTC()); eerr != nil {
		data.IsValid = false
		data.ValidationErrors = append(data.ValidationErrors, "invalid date of expiry")
	} else {
		data.DateOfExpiry = exp
	}

	if !ValidateCheckDigit(passportNumber, passportCD) {
		if corrected, count, err := CorrectCheckDigitField(passportNumber, passportCD); err == nil {
			passportNumber = corrected
			data.PassportNumber = corrected
			data.CorrectionCount += count
			data.ValidationErrors = append(data.ValidationErrors, "passport number check digit corrected")
		} else {
			data.IsValid = false
			data.ValidationErrors = append(data.ValidationErrors, "passport number check digit failed")
		}
	}
	if !ValidateCheckDigit(dobStr, dobCD) {
		if corrected, count, err := CorrectCheckDigitField(dobStr, dobCD); err == nil {
			dobStr = corrected
			data.CorrectionCount += count
			data.ValidationErrors = append(data.ValidationErrors, "date of birth check digit corrected")
		} else {
			data.IsValid = false
			data.ValidationErrors = append(data.ValidationErrors, "date of birth check digit failed")
		}
	}
	if !ValidateCheckDigit(expStr, expCD) {
		if corrected, count, err := CorrectCheckDigitField(expStr, expCD); err == nil {
			expStr = corrected
			data.CorrectionCount += count
			data.ValidationErrors = append(data.ValidationErrors, "date of expiry check digit corrected")
		} else {
			data.IsValid = false
			data.ValidationErrors = append(data.ValidationErrors, "date of expiry check digit failed")
		}
	}
	if !ValidateCheckDigit(optional, optionalCD) {
		if corrected, count, err := CorrectCheckDigitField(optional, optionalCD); err == nil {
			optional = corrected
			data.OptionalData = corrected
			data.CorrectionCount += count
			data.ValidationErrors = append(data.ValidationErrors, "optional data check digit corrected")
		} else {
			data.IsValid = false
			data.ValidationErrors = append(data.ValidationErrors, "optional data check digit failed")
		}
	}

	composite := passportNumber + string(passportCD) + dobStr + string(dobCD) + expStr + string(expCD) + optional + string(optionalCD)
	if !ValidateCheckDigit(composite, compositeCD) {
		if _, count, err := CorrectCheckDigitField(composite, compositeCD); err == nil {
			data.CorrectionCount += count
			data.ValidationErrors = append(data.ValidationErrors, "composite check digit corrected")
		} else {
			data.IsValid = false
			data.ValidationErrors = append(data.ValidationErrors, "composite check digit failed")
		}
	}

	data.Confidence = ComputeConfidence(data)

	return data, nil
}

func CalculateCheckDigit(data string) int {
	weights := []int{7, 3, 1}
	sum := 0
	for i := 0; i < len(data); i++ {
		sum += mrzCharValue(data[i]) * weights[i%3]
	}
	return sum % 10
}

func ValidateCheckDigit(data string, checkDigit byte) bool {
	checkDigit = normalizeDigit(checkDigit)
	if checkDigit < '0' || checkDigit > '9' {
		return false
	}
	return int(checkDigit-'0') == CalculateCheckDigit(data)
}

// CorrectCheckDigitField tries single-character substitutions using
// confusion pairs (O→0, I→1, S→5, B→8, G→6, Z→2) to find a correction
// that passes the check digit. Returns the corrected field or an error
// if the field is already valid or cannot be corrected.
func CorrectCheckDigitField(field string, expectedCheckDigit byte) (string, int, error) {
	if ValidateCheckDigit(field, expectedCheckDigit) {
		return "", 0, fmt.Errorf("field already valid for check digit %c", expectedCheckDigit)
	}
	for i := 0; i < len(field); i++ {
		original := field[i]
		if alternatives, ok := confusionPairs[original]; ok {
			for _, alt := range alternatives {
				corrected := field[:i] + string(alt) + field[i+1:]
				if ValidateCheckDigit(corrected, expectedCheckDigit) {
					return corrected, 1, nil
				}
			}
		}
	}
	return "", 0, fmt.Errorf("unable to correct check digit")
}

func ParseName(nameField string) (surname string, givenNames []string) {
	nameField = strings.TrimSpace(nameField)
	parts := strings.Split(nameField, "<<")
	if len(parts) == 0 {
		return "", nil
	}

	surname = cleanupMRZNamePart(parts[0])
	givenNames = make([]string, 0)
	for _, part := range parts[1:] {
		part = cleanupMRZNamePart(part)
		if part == "" {
			continue
		}
		for _, value := range strings.Fields(part) {
			if value != "" {
				givenNames = append(givenNames, value)
			}
		}
	}
	return surname, givenNames
}

func cleanupMRZNamePart(value string) string {
	value = strings.ReplaceAll(value, "<", " ")
	value = strings.ToUpper(strings.Join(strings.Fields(value), " "))
	return strings.TrimSpace(value)
}

func ParseMRZDate(dateStr string) (time.Time, error) {
	return ParseMRZExpiryDate(dateStr, time.Now().UTC())
}

func ParseMRZBirthDate(dateStr string, asOf time.Time) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if len(dateStr) != 6 {
		return time.Time{}, fmt.Errorf("invalid MRZ date length")
	}

	bytes := []byte(strings.ToUpper(dateStr))
	for i := range bytes {
		bytes[i] = normalizeDigit(bytes[i])
	}
	norm := string(bytes)

	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}

	candidates, err := mrzDateCandidates(norm)
	if err != nil {
		return time.Time{}, err
	}

	var selected time.Time
	for _, candidate := range candidates {
		if candidate.After(asOf) {
			continue
		}
		if selected.IsZero() || candidate.After(selected) {
			selected = candidate
		}
	}
	if !selected.IsZero() {
		return selected, nil
	}

	return candidates[0], nil
}

func ParseMRZExpiryDate(dateStr string, asOf time.Time) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if len(dateStr) != 6 {
		return time.Time{}, fmt.Errorf("invalid MRZ date length")
	}

	bytes := []byte(strings.ToUpper(dateStr))
	for i := range bytes {
		bytes[i] = normalizeDigit(bytes[i])
	}
	norm := string(bytes)

	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}

	candidates, err := mrzDateCandidates(norm)
	if err != nil {
		return time.Time{}, err
	}

	maxFuture := asOf.AddDate(20, 0, 0)
	var selected time.Time
	var selectedDistance time.Duration
	for _, candidate := range candidates {
		if candidate.After(maxFuture) {
			continue
		}
		distance := absoluteDuration(candidate.Sub(asOf))
		if selected.IsZero() || distance < selectedDistance {
			selected = candidate
			selectedDistance = distance
		}
	}
	if !selected.IsZero() {
		return selected, nil
	}

	return candidates[len(candidates)-1], nil
}

func mrzDateCandidates(dateStr string) ([]time.Time, error) {
	yy, err := strconv.Atoi(dateStr[0:2])
	if err != nil {
		return nil, fmt.Errorf("invalid MRZ year: %w", err)
	}
	mm, err := strconv.Atoi(dateStr[2:4])
	if err != nil {
		return nil, fmt.Errorf("invalid MRZ month: %w", err)
	}
	dd, err := strconv.Atoi(dateStr[4:6])
	if err != nil {
		return nil, fmt.Errorf("invalid MRZ day: %w", err)
	}

	candidates := []time.Time{
		time.Date(1900+yy, time.Month(mm), dd, 0, 0, 0, 0, time.UTC),
		time.Date(2000+yy, time.Month(mm), dd, 0, 0, 0, 0, time.UTC),
	}
	for _, candidate := range candidates {
		if int(candidate.Month()) != mm || candidate.Day() != dd {
			return nil, fmt.Errorf("invalid MRZ date value")
		}
	}

	return candidates, nil
}

func absoluteDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}

func CleanMRZLine(line string) (string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", fmt.Errorf("MRZ line is required")
	}

	line = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r':
			return -1
		default:
			return r
		}
	}, line)
	line = strings.ToUpper(line)

	replacements := map[string]string{
		"Â«":  "<",
		"â€¹": "<",
		"â€º": "<",
		"ï¼œ": "<",
		"âŸ¨": "<",
		"âŸ©": "<",
	}
	for from, to := range replacements {
		line = strings.ReplaceAll(line, from, to)
	}

	bytes := []byte(line)
	for i := 0; i < len(bytes); i++ {
		prevIsDigit := i > 0 && bytes[i-1] >= '0' && bytes[i-1] <= '9'
		nextIsDigit := i+1 < len(bytes) && bytes[i+1] >= '0' && bytes[i+1] <= '9'
		if (prevIsDigit && nextIsDigit) && (bytes[i] == 'O' || bytes[i] == 'I') {
			bytes[i] = normalizeDigit(bytes[i])
		}
	}
	line = string(bytes)

	for i := 0; i < len(line); i++ {
		c := line[i]
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '<' {
			continue
		}
		return "", fmt.Errorf("MRZ line contains invalid characters")
	}

	return line, nil
}

func mrzCharValue(c byte) int {
	c = normalizeDigit(byte(strings.ToUpper(string(c))[0]))
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'Z':
		return 10 + int(c-'A')
	case c == '<':
		return 0
	default:
		return 0
	}
}

func normalizeDigit(c byte) byte {
	switch c {
	case 'O':
		return '0'
	case 'I':
		return '1'
	default:
		return c
	}
}

// ValidMRZChars contains all characters allowed in MRZ lines per ICAO 9303.
const ValidMRZChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ<"

// confusionPairs maps characters commonly misread by OCR to their
// likely correct alternatives in the MRZ context.
// Only include substitutions where the OCR-B font glyphs are visually similar.
var confusionPairs = map[byte][]byte{
	'O': {'0'},
	'I': {'1'},
	'S': {'5'},
	'B': {'8'},
	'G': {'6'},
	'Z': {'2'},
}

// ComputeConfidence calculates a confidence score from 0.0 to 1.0
// based on the ratio of passed check digits and corrections applied.
func ComputeConfidence(data *MRZData) float64 {
	if data == nil {
		return 0
	}
	const totalChecks = 5
	failedCount := len(data.ValidationErrors) - data.CorrectionCount
	if failedCount < 0 {
		failedCount = 0
	}
	passed := totalChecks - failedCount
	if passed < 0 {
		passed = 0
	}
	confidence := float64(passed) / float64(totalChecks)
	penalty := float64(data.CorrectionCount) * 0.05
	if confidence-penalty < 0 {
		return 0
	}
	return confidence - penalty
}
