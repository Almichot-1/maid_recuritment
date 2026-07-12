package domain

import "time"

type SelectionStatus string

const (
	SelectionPending  SelectionStatus = "pending"
	SelectionApproved SelectionStatus = "approved"
	SelectionRejected SelectionStatus = "rejected"
	SelectionExpired  SelectionStatus = "expired"
	SelectionReleased SelectionStatus = "released"
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
	WarningSentFlags           int
	ExpiresAt                  time.Time           `gorm:"not null"`
	CreatedAt                  time.Time           `gorm:"not null;default:now()"`
	UpdatedAt                  time.Time           `gorm:"not null;default:now()"`
	Candidate                  *Candidate          `gorm:"foreignKey:CandidateID;references:ID"`
	Progress                   *SelectionProgress  `gorm:"foreignKey:SelectionID;references:ID"`
}

func (Selection) TableName() string {
	return "selections"
}

type SelectionFilters struct {
	SelectedBy    string
	CandidateOwner string
	PairingID     string
	Statuses      []SelectionStatus
	SortBy        string
	SortOrder     string
	Page          int
	PageSize      int
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
	List(filters SelectionFilters) ([]*Selection, error)
	Count(filters SelectionFilters) (int64, error)
	UpdateStatus(id string, status SelectionStatus) error
	GetExpiredSelections() ([]*Selection, error)
}
