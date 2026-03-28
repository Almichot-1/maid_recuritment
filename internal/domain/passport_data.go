package domain

import "time"

type PassportData struct {
	ID                       string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CandidateID              string `gorm:"type:uuid;not null;uniqueIndex"`
	HolderName               string `gorm:"not null"`
	PassportNumber           string `gorm:"not null"`
	CountryCode              string
	Nationality              string    `gorm:"not null"`
	DateOfBirth              time.Time `gorm:"not null"`
	PlaceOfBirth             string
	Gender                   string `gorm:"not null"`
	IssueDate                *time.Time
	ExpiryDate               time.Time `gorm:"not null"`
	IssuingAuthority         string
	MRZLine1                 string    `gorm:"column:mrz_line_1;not null"`
	MRZLine2                 string    `gorm:"column:mrz_line_2;not null"`
	Confidence               float64   `gorm:"not null;default:0"`
	ExtractedAt              time.Time `gorm:"not null;default:now()"`
	PassportWarningSentFlags int       `gorm:"not null;default:0"`
	CreatedAt                time.Time `gorm:"not null;default:now()"`
	UpdatedAt                time.Time `gorm:"not null;default:now()"`
}

func (PassportData) TableName() string {
	return "passport_data"
}

func (p *PassportData) Age(asOf time.Time) int {
	if p == nil || p.DateOfBirth.IsZero() {
		return 0
	}
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	asOf = asOf.UTC()
	dob := p.DateOfBirth.UTC()
	if dob.After(asOf) {
		return 0
	}

	years := asOf.Year() - dob.Year()
	birthdayThisYear := time.Date(asOf.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, time.UTC)
	if asOf.Before(birthdayThisYear) {
		years--
	}
	if years < 0 {
		return 0
	}
	return years
}

type PassportDataRepository interface {
	Upsert(data *PassportData) error
	GetByCandidateID(candidateID string) (*PassportData, error)
	GetExpiringPassports(days int) ([]*PassportData, error)
}
