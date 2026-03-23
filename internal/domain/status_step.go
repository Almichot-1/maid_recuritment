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
	Medical         = "Medical"
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
	VisaProcessing = TicketPending
	FlightBooked   = TicketBooked
	Deployed       = Arrived
)

type StatusStep struct {
	ID          string     `gorm:"type:uuid;primaryKey"`
	CandidateID string     `gorm:"type:uuid;not null;index"`
	StepName    string     `gorm:"not null"`
	StepStatus  StepStatus `gorm:"type:step_status;not null;default:pending"`
	CompletedAt *time.Time
	Notes       string
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
	Update(step *StatusStep) error
}
