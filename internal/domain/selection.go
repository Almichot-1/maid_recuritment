package domain

import "time"

type SelectionStatus string

const (
	SelectionPending  SelectionStatus = "pending"
	SelectionApproved SelectionStatus = "approved"
	SelectionRejected SelectionStatus = "rejected"
	SelectionExpired  SelectionStatus = "expired"
)

type Selection struct {
	ID                         string          `gorm:"type:uuid;primaryKey"`
	CandidateID                string          `gorm:"type:uuid;not null;index"`
	PairingID                  string          `gorm:"type:uuid;not null;index"`
	SelectedBy                 string          `gorm:"type:uuid;not null;index"`
	Status                     SelectionStatus `gorm:"type:selection_status;not null;default:pending"`
	EmployerContractURL        string
	EmployerContractFileName   string
	EmployerContractUploadedAt *time.Time
	EmployerIDURL              string
	EmployerIDFileName         string
	EmployerIDUploadedAt       *time.Time
	ExpiresAt                  time.Time `gorm:"not null"`
	CreatedAt                  time.Time `gorm:"not null;default:now()"`
	UpdatedAt                  time.Time `gorm:"not null;default:now()"`
}

func (Selection) TableName() string {
	return "selections"
}

type SelectionRepository interface {
	Create(selection *Selection) error
	GetByID(id string) (*Selection, error)
	GetByCandidateID(candidateID string) (*Selection, error)
	GetByCandidateIDAndPairingID(candidateID, pairingID string) (*Selection, error)
	GetBySelectedBy(userID string) ([]*Selection, error)
	GetBySelectedByAndPairing(userID, pairingID string) ([]*Selection, error)
	GetByCandidateOwner(userID string) ([]*Selection, error)
	GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*Selection, error)
	UpdateStatus(id string, status SelectionStatus) error
	GetExpiredSelections() ([]*Selection, error)
}
