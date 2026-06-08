package domain

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type CandidateStatus string

const (
	CandidateStatusDraft       CandidateStatus = "draft"
	CandidateStatusAvailable   CandidateStatus = "available"
	CandidateStatusLocked      CandidateStatus = "locked"
	CandidateStatusUnderReview CandidateStatus = "under_review"
	CandidateStatusApproved    CandidateStatus = "approved"
	CandidateStatusRejected    CandidateStatus = "rejected"
	CandidateStatusInProgress  CandidateStatus = "in_progress"
	CandidateStatusCompleted   CandidateStatus = "completed"
)

type Candidate struct {
	ID                   string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedBy            string `gorm:"type:uuid;not null"`
	FullName             string `gorm:"not null"`
	Nationality          string
	DateOfBirth          *time.Time `gorm:"type:date"`
	Age                  *int
	PlaceOfBirth         string
	Religion             string
	MaritalStatus        string
	ChildrenCount        *int
	EducationLevel       string
	ExperienceYears      *int
	CountryOfExperience  string
	CountryApplied       string `gorm:"column:country_applied"`
	SalaryOffered        string `gorm:"column:salary_offered"`
	Languages            json.RawMessage `gorm:"type:jsonb;not null;default:'[]'::jsonb"`
	Skills               json.RawMessage `gorm:"type:jsonb;not null;default:'[]'::jsonb"`
	Status               CandidateStatus `gorm:"type:candidate_status;not null;default:draft"`
	LockedBy             *string         `gorm:"type:uuid"`
	LockedAt             *time.Time
	LockExpiresAt        *time.Time
	CVPDFURL             string              `gorm:"column:cv_pdf_url"`
	CreatedAt            time.Time           `gorm:"not null;default:now()"`
	UpdatedAt            time.Time           `gorm:"not null;default:now()"`
	DeletedAt            gorm.DeletedAt      `gorm:"index"`
	Documents            []CandidateDocument `gorm:"foreignKey:CandidateID"`
}

func (Candidate) TableName() string {
	return "candidates"
}

type CandidateDocument struct {
	ID           string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CandidateID  string `gorm:"type:uuid;not null"`
	DocumentType string `gorm:"type:document_type;not null"`
	FileURL      string `gorm:"not null"`
	FileName     string
	FileSize     *int64
	UploadedAt   time.Time `gorm:"not null;default:now()"`
}

func (CandidateDocument) TableName() string {
	return "documents"
}

type CandidateFilters struct {
	Statuses      []CandidateStatus
	MinAge        *int
	MaxAge        *int
	MinExperience *int
	MaxExperience *int
	Languages     []string
	Search        string
	CreatedBy     string
	PairingID     string
	SharedOnly    bool
	Page          int
	PageSize      int
}

type CandidateRepository interface {
	Create(candidate *Candidate) error
	GetByID(id string) (*Candidate, error)
	// GetByIDLean fetches only the columns needed for ownership and status
	// checks (id, created_by, status, locked_by, lock_expires_at). It does NOT
	// preload Documents, so it is significantly cheaper than GetByID.
	GetByIDLean(id string) (*Candidate, error)
	List(filters CandidateFilters) ([]*Candidate, error)
	Update(candidate *Candidate) error
	Delete(id string) error
	Lock(candidateID, lockedBy string, expiresAt time.Time) error
	Unlock(candidateID string) error
}
