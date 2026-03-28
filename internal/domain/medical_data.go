package domain

import "time"

type MedicalData struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CandidateID      string    `gorm:"type:uuid;not null;uniqueIndex"`
	DocumentID       string    `gorm:"type:uuid;not null"`
	ExpiryDate       time.Time `gorm:"not null"`
	RawText          string
	ExtractedAt      time.Time `gorm:"not null;default:now()"`
	WarningSentFlags int       `gorm:"not null;default:0"`
	CreatedAt        time.Time `gorm:"not null;default:now()"`
	UpdatedAt        time.Time `gorm:"not null;default:now()"`
}

func (MedicalData) TableName() string {
	return "medical_data"
}

type MedicalDataRepository interface {
	Upsert(data *MedicalData) error
	GetByCandidateID(candidateID string) (*MedicalData, error)
	GetExpiringMedical(days int) ([]*MedicalData, error)
}
