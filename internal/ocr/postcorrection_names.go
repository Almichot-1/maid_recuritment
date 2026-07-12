package ocr

import (
	"fmt"
	"strings"
)

// validateNames checks surname and given names for common OCR issues.
// Returns a list of warning strings (nil if all checks pass).
func validateNames(surname string, givenNames []string) []string {
	var warnings []string

	// 1. Surname checks
	if strings.TrimSpace(surname) == "" {
		warnings = append(warnings, "surname is empty")
	} else if len(surname) < 2 {
		warnings = append(warnings, "surname is too short (single character)")
	} else if isAllSameCharacter(surname) {
		warnings = append(warnings, "surname appears to be invalid (all same character)")
	} else if strings.Contains(surname, "<") {
		warnings = append(warnings, "surname contains filler characters")
	}

	// 2. Given names checks
	if len(givenNames) == 0 || (len(givenNames) == 1 && strings.TrimSpace(givenNames[0]) == "") {
		warnings = append(warnings, "given names are empty")
	} else {
		for i, name := range givenNames {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if len(name) == 1 {
				warnings = append(warnings, fmt.Sprintf("given name %q appears to be a single initial (potential OCR issue)", name))
			}
			if isAllSameCharacter(name) && len(name) > 1 {
				warnings = append(warnings, fmt.Sprintf("given name %q appears to be invalid (all same character)", name))
			}
			_ = i
		}
	}

	// 3. Surname equals given name (unusual but possible)
	if surname != "" && len(givenNames) > 0 {
		for _, name := range givenNames {
			if strings.EqualFold(surname, name) {
				warnings = append(warnings, "surname and given name are identical")
				break
			}
		}
	}

	return warnings
}

// isAllSameCharacter checks if a string consists entirely of the same character.
func isAllSameCharacter(s string) bool {
	if len(s) == 0 {
		return false
	}
	first := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != first {
			return false
		}
	}
	return true
}

// correctNameOcr attempts to fix common OCR confusions in name fields.
// Unlike MRZ fields, name fields don't have check digits, so O→0 and I→1
// substitutions may go undetected. This function applies corrections
// where they're clearly wrong (e.g., digits in a name field).
// Returns the corrected name and true if changes were made.
func correctNameOcr(name string) (string, bool) {
	if name == "" {
		return name, false
	}

	corrected := []byte(name)
	changed := false

	for i, c := range corrected {
		switch c {
		case '0':
			corrected[i] = 'O'
			changed = true
		case '1':
			corrected[i] = 'I'
			changed = true
		case '5':
			corrected[i] = 'S'
			changed = true
		case '8':
			corrected[i] = 'B'
			changed = true
		case '6':
			corrected[i] = 'G'
			changed = true
		case '2':
			corrected[i] = 'Z'
			changed = true
		}
	}

	if changed {
		return string(corrected), true
	}
	return name, false
}

// validatePassportNames runs all name checks on the PassportData struct.
func validatePassportNames(data *PassportData) []string {
	if data == nil {
		return nil
	}
	given := strings.Fields(data.GivenNames)
	return validateNames(data.Surname, given)
}
