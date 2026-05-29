package domain

import "time"

// MedicalData stores metadata for a candidate's uploaded medical certificate.
type MedicalData struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CandidateID string     `gorm:"type:uuid;not null;uniqueIndex"`
	DocumentID  string     `gorm:"type:uuid;not null"`
	IssueDate   *time.Time `gorm:"type:date"`
	ExpiryDate  *time.Time `gorm:"type:date"`
	ExtractedAt time.Time  `gorm:"not null;default:now()"`
	CreatedAt   time.Time  `gorm:"not null;default:now()"`
	UpdatedAt   time.Time  `gorm:"not null;default:now()"`
}

func (MedicalData) TableName() string {
	return "medical_data"
}

// MedicalDataRepository persists extracted medical certificate metadata.
type MedicalDataRepository interface {
	Upsert(data *MedicalData) error
	GetByCandidateID(candidateID string) (*MedicalData, error)
}
