package domain

import "time"

type DocumentType string

const (
	Passport DocumentType = "passport"
	Photo    DocumentType = "photo"
	Video    DocumentType = "video"
)

type Document struct {
	ID           string       `gorm:"type:uuid;primaryKey"`
	CandidateID  string       `gorm:"type:uuid;not null;index"`
	DocumentType DocumentType `gorm:"type:document_type;not null"`
	FileURL      string       `gorm:"not null"`
	FileName     string
	FileSize     int64
	UploadedAt   time.Time `gorm:"not null;default:now()"`
}

func (Document) TableName() string {
	return "documents"
}

type DocumentRepository interface {
	Create(document *Document) error
	GetByCandidateID(candidateID string) ([]*Document, error)
	Delete(id string) error
}
