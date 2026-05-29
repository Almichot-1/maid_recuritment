package domain

import "time"

// PassportWarning bitmask constants – track which passport-expiry alerts have
// already been sent so each level fires only once.
// Bit 0 = 6-month warning, Bit 1 = 3-month warning, Bit 2 = 1-month warning.
const (
	PassportWarning6Months = 1 << 0 // 1
	PassportWarning3Months = 1 << 1 // 2
	PassportWarning1Month  = 1 << 2 // 4
)

// PassportData holds all fields extracted from a candidate's passport via OCR.
// There is at most one PassportData row per candidate (enforced by a unique
// index on candidate_id in the passport_data table).
type PassportData struct {
	ID                       string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CandidateID              string     `gorm:"type:uuid;not null;uniqueIndex"`
	HolderName               string     `gorm:"not null;default:''"`
	PassportNumber           string     `gorm:"not null;default:''"`
	Nationality              string     `gorm:"not null;default:''"`
	CountryCode              string     `gorm:"not null;default:''"`
	DateOfBirth              *time.Time `gorm:"type:date"`
	PlaceOfBirth             string     `gorm:"not null;default:''"`
	Gender                   string     `gorm:"not null;default:''"`
	IssueDate                *time.Time `gorm:"type:date"`
	ExpiryDate               *time.Time `gorm:"type:date"`
	PlaceOfIssue             string     `gorm:"not null;default:''"`
	IssuingAuthority         string     `gorm:"not null;default:''"`
	MRZLine1                 string     `gorm:"not null;default:''"`
	MRZLine2                 string     `gorm:"not null;default:''"`
	Confidence               float64    `gorm:"not null;default:0"`
	PassportWarningSentFlags int        `gorm:"not null;default:0"`
	ExtractedAt              time.Time  `gorm:"not null;default:now()"`
	CreatedAt                time.Time  `gorm:"not null;default:now()"`
	UpdatedAt                time.Time  `gorm:"not null;default:now()"`
}

func (PassportData) TableName() string {
	return "passport_data"
}

// IsExpired returns true when the passport expiry date is in the past.
func (p *PassportData) IsExpired() bool {
	if p == nil || p.ExpiryDate == nil {
		return false
	}
	return p.ExpiryDate.Before(time.Now().UTC())
}

// DaysUntilExpiry returns the number of whole days until the passport expires.
// Returns -1 when no expiry date is set.
func (p *PassportData) DaysUntilExpiry() int {
	if p == nil || p.ExpiryDate == nil {
		return -1
	}
	diff := time.Until(*p.ExpiryDate)
	if diff < 0 {
		return int(diff.Hours() / 24) // negative = already expired
	}
	return int(diff.Hours() / 24)
}

// Age computes the holder's age from DateOfBirth as of today.
// Returns 0 when DateOfBirth is not set.
func (p *PassportData) Age() int {
	if p == nil || p.DateOfBirth == nil {
		return 0
	}
	now := time.Now().UTC()
	dob := p.DateOfBirth.UTC()
	years := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		years--
	}
	if years < 0 {
		return 0
	}
	return years
}

// PassportDataRepository defines the persistence operations for PassportData.
type PassportDataRepository interface {
	// Upsert inserts a new PassportData row or updates the existing one for the
	// same candidate_id.
	Upsert(data *PassportData) error

	// GetByCandidateID returns the PassportData for the given candidate, or
	// ErrPassportDataNotFound when none exists.
	GetByCandidateID(candidateID string) (*PassportData, error)

	// GetExpiringPassports returns PassportData rows whose expiry_date falls
	// within the next daysAhead days and whose passport_warning_sent_flags does
	// NOT yet have flagBit set.
	GetExpiringPassports(daysAhead int, flagBit int) ([]*PassportData, error)

	// UpdateWarningSentFlags sets the passport_warning_sent_flags column for
	// the given PassportData ID.
	UpdateWarningSentFlags(id string, flags int) error
}
