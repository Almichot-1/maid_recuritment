package ocr

// postCorrectPassportData runs all post-correction steps on parsed PassportData.
// It corrects OCR errors in names, validates/corrects country codes and
// nationalities, and applies confidence penalties for unresolvable issues.
// Returns nil if data is nil, otherwise returns the modified data pointer.
func postCorrectPassportData(data *PassportData) *PassportData {
	if data == nil {
		return nil
	}

	if corrected, changed := correctNameOcr(data.Surname); changed {
		data.Surname = corrected
	}

	if corrected, changed := correctNameOcr(data.GivenNames); changed {
		data.GivenNames = corrected
	}

	if data.CountryCode != "" && !isValidCountryCode(data.CountryCode) {
		if corrected, changed := correctCountryCode(data.CountryCode); changed {
			data.CountryCode = corrected
		} else {
			data.Confidence *= 0.9
		}
	}

	if data.Nationality != "" && !isValidCountryCode(data.Nationality) {
		if corrected, changed := correctNationality(data.Nationality); changed {
			data.Nationality = corrected
		}
	}

	return data
}
