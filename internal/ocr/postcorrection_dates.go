package ocr

import "time"

// validateDateConsistency checks cross-field date consistency for passport dates.
// Checks (only for non-zero date pairs):
// 1. DateOfBirth < DateOfIssue
// 2. DateOfIssue < DateOfExpiry
// 3. DateOfBirth < DateOfExpiry
// 4. DateOfIssue not > 7 days in the future
// 5. Expiry not more than 20 years after issue
// 6. Birth not more than 120 years ago
// 7. Birth not in the future
// Returns nil if all checks pass, otherwise warning strings.
func validateDateConsistency(dob, issue, expiry time.Time) []string {
	now := time.Now().UTC()
	var warnings []string

	if !dob.IsZero() && !issue.IsZero() && !issue.After(dob) {
		warnings = append(warnings, "date of issue must be after date of birth")
	}
	if !issue.IsZero() && !expiry.IsZero() && !expiry.After(issue) {
		warnings = append(warnings, "date of expiry must be after date of issue")
	}
	if !dob.IsZero() && !expiry.IsZero() && !expiry.After(dob) {
		warnings = append(warnings, "date of expiry must be after date of birth")
	}
	if !issue.IsZero() && issue.After(now.AddDate(0, 0, 7)) {
		warnings = append(warnings, "date of issue is in the future")
	}
	if !issue.IsZero() && !expiry.IsZero() && expiry.After(issue.AddDate(20, 0, 0)) {
		warnings = append(warnings, "date of expiry exceeds maximum passport validity (20 years)")
	}
	if !dob.IsZero() && dob.Before(now.AddDate(-120, 0, 0)) {
		warnings = append(warnings, "date of birth is unreasonably far in the past")
	}
	if !dob.IsZero() && dob.After(now) {
		warnings = append(warnings, "date of birth is in the future")
	}
	if warnings == nil {
		return nil
	}
	return warnings
}

// validatePassportDates runs all date checks on the PassportData struct.
func validatePassportDates(data *PassportData) []string {
	if data == nil {
		return nil
	}
	return validateDateConsistency(data.DateOfBirth, data.DateOfIssue, data.DateOfExpiry)
}
