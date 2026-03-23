package domain

import "time"

type AgencyPairingStatus string

const (
	AgencyPairingActive    AgencyPairingStatus = "active"
	AgencyPairingSuspended AgencyPairingStatus = "suspended"
	AgencyPairingEnded     AgencyPairingStatus = "ended"
)

type AgencyPairing struct {
	ID                string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EthiopianUserID   string              `gorm:"type:uuid;not null;index"`
	ForeignUserID     string              `gorm:"type:uuid;not null;index"`
	Status            AgencyPairingStatus `gorm:"type:agency_pairing_status;not null;default:active"`
	ApprovedByAdminID *string             `gorm:"type:uuid"`
	ApprovedAt        *time.Time
	EndedAt           *time.Time
	Notes             *string
	CreatedAt         time.Time `gorm:"not null;default:now()"`
	UpdatedAt         time.Time `gorm:"not null;default:now()"`
}

func (AgencyPairing) TableName() string {
	return "agency_pairings"
}

type AgencyPairingFilters struct {
	UserID          string
	EthiopianUserID string
	ForeignUserID   string
	Status          *AgencyPairingStatus
}

type CandidatePairShare struct {
	ID             string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PairingID      string `gorm:"type:uuid;not null;index"`
	CandidateID    string `gorm:"type:uuid;not null;index"`
	SharedByUserID string `gorm:"type:uuid;not null"`
	IsActive       bool   `gorm:"not null;default:true"`
	SharedAt       time.Time
	UnsharedAt     *time.Time
	CreatedAt      time.Time `gorm:"not null;default:now()"`
	UpdatedAt      time.Time `gorm:"not null;default:now()"`
}

func (CandidatePairShare) TableName() string {
	return "candidate_pair_shares"
}

type AgencyPairingRepository interface {
	Create(pairing *AgencyPairing) error
	GetByID(id string) (*AgencyPairing, error)
	GetActiveByUsers(ethiopianUserID, foreignUserID string) (*AgencyPairing, error)
	List(filters AgencyPairingFilters) ([]*AgencyPairing, error)
	Update(pairing *AgencyPairing) error
}

type CandidatePairShareRepository interface {
	Create(share *CandidatePairShare) error
	GetActiveByPairingAndCandidate(pairingID, candidateID string) (*CandidatePairShare, error)
	ListByCandidateID(candidateID string, activeOnly bool) ([]*CandidatePairShare, error)
	ListByPairingID(pairingID string, activeOnly bool) ([]*CandidatePairShare, error)
	Deactivate(pairingID, candidateID string, unsharedAt time.Time) error
}
