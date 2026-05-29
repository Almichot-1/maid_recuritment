package domain

import "time"

type SelectionStatus string

const (
	SelectionPending  SelectionStatus = "pending"
	SelectionApproved SelectionStatus = "approved"
	SelectionRejected SelectionStatus = "rejected"
	SelectionExpired  SelectionStatus = "expired"
)

// WarningSentFlags bitmask constants for selection expiry warnings.
// Bit 0 = 24 h warning sent, Bit 1 = 6 h warning sent, Bit 2 = 1 h warning sent.
const (
	WarningSent24h = 1 << 0 // 1
	WarningSent6h  = 1 << 1 // 2
	WarningSent1h  = 1 << 2 // 4
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
	// WarningSentFlags is a bitmask tracking which pre-expiry warnings have
	// already been dispatched for this selection so each warning fires only once.
	// Use the WarningSent24h / WarningSent6h / WarningSent1h constants.
	WarningSentFlags int       `gorm:"not null;default:0"`
	CreatedAt        time.Time `gorm:"not null;default:now()"`
	UpdatedAt        time.Time `gorm:"not null;default:now()"`
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
	UpdateWarningSentFlags(id string, flags int) error
	GetExpiredSelections() ([]*Selection, error)
	// GetPendingExpiringSoon returns pending selections whose expires_at falls
	// within the next windowHours hours and whose warning_sent_flags does NOT
	// yet have the provided flagBit set.
	GetPendingExpiringSoon(windowHours int, flagBit int) ([]*Selection, error)
}
