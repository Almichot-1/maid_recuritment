package ocr

import "strings"

// isoCountryCodes maps ISO 3166-1 alpha-3 codes to full country names.
// Includes the most common passport-issuing countries.
var isoCountryCodes = map[string]string{
	"ARE": "UNITED ARAB EMIRATES",
	"ARG": "ARGENTINA",
	"AUS": "AUSTRALIA",
	"AUT": "AUSTRIA",
	"BEL": "BELGIUM",
	"BGD": "BANGLADESH",
	"BGR": "BULGARIA",
	"BRA": "BRAZIL",
	"CAN": "CANADA",
	"CHE": "SWITZERLAND",
	"CHN": "CHINA",
	"COL": "COLOMBIA",
	"CYP": "CYPRUS",
	"CZE": "CZECH REPUBLIC",
	"DEU": "GERMANY",
	"DNK": "DENMARK",
	"DZA": "ALGERIA",
	"EGY": "EGYPT",
	"ESP": "SPAIN",
	"ETH": "ETHIOPIA",
	"FIN": "FINLAND",
	"FRA": "FRANCE",
	"GBR": "UNITED KINGDOM",
	"GHA": "GHANA",
	"GRC": "GREECE",
	"HKG": "HONG KONG",
	"HUN": "HUNGARY",
	"IDN": "INDONESIA",
	"IND": "INDIA",
	"IRL": "IRELAND",
	"IRN": "IRAN",
	"IRQ": "IRAQ",
	"ISL": "ICELAND",
	"ISR": "ISRAEL",
	"ITA": "ITALY",
	"JOR": "JORDAN",
	"JPN": "JAPAN",
	"KEN": "KENYA",
	"KHM": "CAMBODIA",
	"KOR": "SOUTH KOREA",
	"KWT": "KUWAIT",
	"LBN": "LEBANON",
	"LKA": "SRI LANKA",
	"LTU": "LITHUANIA",
	"LUX": "LUXEMBOURG",
	"LVA": "LATVIA",
	"MAR": "MOROCCO",
	"MEX": "MEXICO",
	"MKD": "NORTH MACEDONIA",
	"MMR": "MYANMAR",
	"MYS": "MALAYSIA",
	"NGA": "NIGERIA",
	"NLD": "NETHERLANDS",
	"NOR": "NORWAY",
	"NZL": "NEW ZEALAND",
	"OMN": "OMAN",
	"PAK": "PAKISTAN",
	"PER": "PERU",
	"PHL": "PHILIPPINES",
	"POL": "POLAND",
	"PRT": "PORTUGAL",
	"QAT": "QATAR",
	"ROU": "ROMANIA",
	"RUS": "RUSSIA",
	"SAU": "SAUDI ARABIA",
	"SGP": "SINGAPORE",
	"SRB": "SERBIA",
	"SVK": "SLOVAKIA",
	"SVN": "SLOVENIA",
	"SWE": "SWEDEN",
	"THA": "THAILAND",
	"TUN": "TUNISIA",
	"TUR": "TURKEY",
	"TWN": "TAIWAN",
	"UKR": "UKRAINE",
	"USA": "UNITED STATES",
	"VNM": "VIETNAM",
	"ZAF": "SOUTH AFRICA",
	"ZWE": "ZIMBABWE",
}

// commonCountryConfusions maps known OCR confusions for country codes.
// These are hard-coded corrections for frequently misread codes.
var commonCountryConfusions = map[string]string{
	"UT0": "UTO",
	"UTO": "UTO",
}

// GetISOCountryCodes returns the ISO 3166-1 alpha-3 map (for testing).
func GetISOCountryCodes() map[string]string {
	result := make(map[string]string, len(isoCountryCodes))
	for k, v := range isoCountryCodes {
		result[k] = v
	}
	return result
}

// isValidCountryCode checks if the given 3-letter code is a known ISO 3166-1 alpha-3 code.
func isValidCountryCode(code string) bool {
	_, ok := isoCountryCodes[code]
	return ok
}

// correctCountryCode attempts to correct a potentially mis-OCR'd country code.
// Returns the corrected code (from isoCountryCodes) and true if correction was applied.
// Uses Levenshtein distance to find the closest match.
// Only corrects if the closest match has similarity >= 0.7 (same threshold as fuzzy label matching).
// If the code is already valid, returns (code, false).
func correctCountryCode(code string) (string, bool) {
	code = strings.TrimSpace(code)
	code = strings.ToUpper(code)

	if len(code) != 3 {
		if corrected, ok := commonCountryConfusions[code]; ok {
			return corrected, true
		}
		return code, false
	}

	if isValidCountryCode(code) {
		return code, false
	}

	if corrected, ok := commonCountryConfusions[code]; ok {
		return corrected, true
	}

	bestMatch := ""
	bestDist := 3

	for isoCode := range isoCountryCodes {
		dist := levenshteinDistance(code, isoCode)
		if dist < bestDist {
			bestDist = dist
			bestMatch = isoCode
		}
	}

	if bestMatch != "" {
		similarity := 1.0 - float64(bestDist)/3.0
		if similarity >= 0.7 {
			return bestMatch, true
		}
	}

	return code, false
}

// correctNationality attempts to correct a potentially mis-OCR'd nationality field.
// Nationality fields are typically the same as the country code.
// Returns the corrected value and true if correction was applied.
func correctNationality(nationality string) (string, bool) {
	return correctCountryCode(nationality)
}
