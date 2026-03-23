package domain

import "time"

type ApprovalDecision string

const (
	ApprovalApproved ApprovalDecision = "approved"
	ApprovalRejected ApprovalDecision = "rejected"
)

type Approval struct {
	ID          string           `gorm:"type:uuid;primaryKey"`
	SelectionID string           `gorm:"type:uuid;not null;index"`
	UserID      string           `gorm:"type:uuid;not null;index"`
	Decision    ApprovalDecision `gorm:"type:approval_decision;not null"`
	DecidedAt   time.Time        `gorm:"not null;default:now()"`
	User        *User            `gorm:"foreignKey:UserID;references:ID"`
}

func (Approval) TableName() string {
	return "approvals"
}

type ApprovalRepository interface {
	Create(approval *Approval) error
	GetBySelectionID(selectionID string) ([]*Approval, error)
	GetBySelectionAndUser(selectionID, userID string) (*Approval, error)
}
