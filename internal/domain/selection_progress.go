package domain

import "time"

// SelectionProgress tracks the recruitment progress for a selection
type SelectionProgress struct {
	ID          string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	SelectionID string     `gorm:"type:uuid;not null;uniqueIndex:uq_selection_progress_selection_id" json:"selection_id"`
	
	// COC (Certificate of Competency)
	COCStatus             string     `gorm:"type:text;not null;default:'pending'" json:"coc_status"`
	COCType               *string    `gorm:"type:text" json:"coc_type,omitempty"`
	COCDocumentURL        string     `gorm:"type:text;column:coc_document_url" json:"-"`
	COCDocumentFileName   string     `gorm:"type:text;column:coc_document_name" json:"coc_document_file_name,omitempty"`
	COCDocumentUploadedAt *time.Time `gorm:"column:coc_uploaded_at" json:"coc_document_uploaded_at,omitempty"`
	
	// Medical
	MedicalStatus             string     `gorm:"type:text;not null;default:'pending'" json:"medical_status"`
	MedicalDocumentURL        string     `gorm:"type:text;column:medical_document_url" json:"-"`
	MedicalDocumentFileName   string     `gorm:"type:text;column:medical_document_name" json:"medical_document_file_name,omitempty"`
	MedicalDocumentUploadedAt *time.Time `gorm:"column:medical_uploaded_at" json:"medical_document_uploaded_at,omitempty"`
	
	// Visa
	VisaStatus             string     `gorm:"type:text;not null;default:'pending'" json:"visa_status"`
	VisaDocumentURL        string     `gorm:"type:text;column:visa_document_url" json:"-"`
	VisaDocumentFileName   string     `gorm:"type:text;column:visa_document_name" json:"visa_document_file_name,omitempty"`
	VisaDocumentUploadedAt *time.Time `gorm:"column:visa_uploaded_at" json:"visa_document_uploaded_at,omitempty"`
	
	// Ticket
	TicketStatus             string     `gorm:"type:text;not null;default:'pending'" json:"ticket_status"`
	TicketDocumentURL        string     `gorm:"type:text;column:ticket_document_url" json:"-"`
	TicketDocumentFileName   string     `gorm:"type:text;column:ticket_document_name" json:"ticket_document_file_name,omitempty"`
	TicketDocumentUploadedAt *time.Time `gorm:"column:ticket_uploaded_at" json:"ticket_document_uploaded_at,omitempty"`
	
	// Arrival
	ArrivalStatus         string     `gorm:"type:text;not null;default:'not_arrived'" json:"arrival_status"`
	ArrivalDate           *time.Time `gorm:"type:date" json:"arrival_date,omitempty"`
	ArrivalCity           *string    `gorm:"type:text" json:"arrival_city,omitempty"`
	DestinationCountry    *string    `gorm:"type:text" json:"destination_country,omitempty"`
	DepartureDate         *time.Time `gorm:"type:date" json:"departure_date,omitempty"`
	ArrivalDocumentURL    string     `gorm:"type:text;column:arrival_document_url" json:"-"`
	ArrivalDocumentFileName string   `gorm:"type:text;column:arrival_document_name" json:"arrival_document_file_name,omitempty"`
	ArrivalDocumentUploadedAt *time.Time `gorm:"column:arrival_uploaded_at" json:"arrival_document_uploaded_at,omitempty"`

	Notes       string     `gorm:"type:text" json:"notes,omitempty"`
	
	// Metadata
	UpdatedBy string    `gorm:"type:uuid;not null" json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SelectionProgress) TableName() string {
	return "selection_progress"
}

// SelectionProgressRepository defines the interface for progress data access
type SelectionProgressRepository interface {
	Create(progress *SelectionProgress) error
	GetByID(id string) (*SelectionProgress, error)
	GetBySelectionID(selectionID string) (*SelectionProgress, error)
	Update(progress *SelectionProgress) error
}

// Valid status values - Generic
const (
	ProgressStatusPending    = "pending"
	ProgressStatusInProgress = "in_progress"
	ProgressStatusDone       = "done"
	ProgressStatusFailed     = "failed"
)

// Valid status values
const (
	// COC statuses
	COCStatusPending    = "pending"
	COCStatusInProgress = "in_progress"
	COCStatusDone       = "done"
	COCStatusFailed     = "failed"
	
	// COC types
	COCTypeOnline  = "online"
	COCTypeOffline = "offline"
	
	// Medical statuses
	MedicalStatusPending    = "pending"
	MedicalStatusInProgress = "in_progress"
	MedicalStatusDone       = "done"
	MedicalStatusFailed     = "failed"
	
	// Visa statuses
	VisaStatusPending    = "pending"
	VisaStatusInProgress = "in_progress"
	VisaStatusApproved   = "approved"
	VisaStatusRejected   = "rejected"
	
	// Ticket statuses
	TicketStatusPending   = "pending"
	TicketStatusBooked    = "booked"
	TicketStatusConfirmed = "confirmed"
	TicketStatusArrived   = "arrived"
	
	// Arrival statuses
	ArrivalStatusNotArrived = "not_arrived"
	ArrivalStatusInTransit  = "in_transit"
	ArrivalStatusArrived    = "arrived"
)
