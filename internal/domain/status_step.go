package domain

import "time"

type StepStatus string

const (
	Pending    StepStatus = "pending"
	InProgress StepStatus = "in_progress"
	Completed  StepStatus = "completed"
	Failed     StepStatus = "failed"
)

const (
	Medical      = "Medical"
	CoC          = "CoC"
	Visa         = "Visa"
	Ticket       = "Ticket"
	ArrivalCity  = "Arrival City"

	// Legacy step names kept for backward compatibility
	CoCPending      = "CoC Pending"
	CoCOnline       = "CoC Online"
	LMISPending     = "LMIS Pending"
	LMISIssued      = "LMIS Issued"
	TicketPending   = "Ticket Pending"
	TicketBooked    = "Ticket Booked"
	TicketConfirmed = "Ticket Confirmed"
	Arrived         = "Arrived"

	MedicalTest    = Medical
	LMISApproval   = LMISIssued
	VisaProcessing = Visa
	FlightBooked   = TicketBooked
	Deployed       = Arrived
)

const (
	CoCNotOnline = "not_online"
	CoCOnlineStatus = "online"
)

type StatusStep struct {
	ID          string     `gorm:"type:uuid;primaryKey"`
	CandidateID string     `gorm:"type:uuid;not null;index"`
	StepName    string     `gorm:"not null"`
	StepStatus  StepStatus `gorm:"type:step_status;not null;default:pending"`
	CompletedAt *time.Time
	Notes       string
	CoCStatus   *string `gorm:"type:text"`
	ArrivalCity *string `gorm:"type:text"`
	UpdatedBy   string    `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
}

func (StatusStep) TableName() string {
	return "status_steps"
}

type StatusStepRepository interface {
	Create(step *StatusStep) error
	GetByCandidateID(candidateID string) ([]*StatusStep, error)
	GetByCandidateIDs(candidateIDs []string) ([]*StatusStep, error)
	GetByCandidateIDAndStepName(candidateID, stepName string) (*StatusStep, error)
	Update(step *StatusStep) error
}
