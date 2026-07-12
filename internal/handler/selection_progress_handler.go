package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

// progressPercentageSteps defines the steps counted for progress percentage
var progressPercentageSteps = []struct {
	getStatus func(*domain.SelectionProgress) string
}{
	{func(p *domain.SelectionProgress) string { return p.COCStatus }},
	{func(p *domain.SelectionProgress) string { return p.MedicalStatus }},
	{func(p *domain.SelectionProgress) string { return p.VisaStatus }},
	{func(p *domain.SelectionProgress) string { return p.TicketStatus }},
	{func(p *domain.SelectionProgress) string { return p.ArrivalStatus }},
}

func computeProgressPercentage(p *domain.SelectionProgress) float64 {
	if p == nil {
		return 0
	}
	completed := 0
	for _, step := range progressPercentageSteps {
		if service.IsProgressStepCompleted(step.getStatus(p)) {
			completed++
		}
	}
	return float64(completed) / float64(len(progressPercentageSteps)) * 100
}

type SelectionProgressResponse struct {
	ID         string `json:"id"`
	SelectionID string `json:"selection_id"`
	UpdatedBy  string `json:"updated_by"`

	// COC fields
	COCStatus           string                    `json:"coc_status"`
	COCType             *string                   `json:"coc_type,omitempty"`
	COCDocument         *SelectionDocumentSummary `json:"coc_document,omitempty"`

	// Medical fields
	MedicalStatus   string                    `json:"medical_status"`
	MedicalDocument *SelectionDocumentSummary `json:"medical_document,omitempty"`

	// Visa fields
	VisaStatus   string                    `json:"visa_status"`
	VisaDocument *SelectionDocumentSummary `json:"visa_document,omitempty"`

	// Ticket fields
	TicketStatus   string                    `json:"ticket_status"`
	TicketDocument *SelectionDocumentSummary `json:"ticket_document,omitempty"`

	// Arrival fields
	ArrivalStatus         string                    `json:"arrival_status"`
	ArrivalDate           *string                   `json:"arrival_date,omitempty"`
	ArrivalCity           *string                   `json:"arrival_city,omitempty"`
	DestinationCountry    *string                   `json:"destination_country,omitempty"`
	DepartureDate         *string                   `json:"departure_date,omitempty"`
	ArrivalDocument       *SelectionDocumentSummary `json:"arrival_document,omitempty"`
	Notes                 string                    `json:"notes,omitempty"`

	ProgressPercentage float64 `json:"progress_percentage"`
	UpdatedByName      string  `json:"updated_by_name"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UpdateProgressRequest struct {
	COCStatus     *string `json:"coc_status,omitempty"`
	COCType       *string `json:"coc_type,omitempty"`
	MedicalStatus *string `json:"medical_status,omitempty"`
	VisaStatus    *string `json:"visa_status,omitempty"`
	TicketStatus  *string `json:"ticket_status,omitempty"`
	ArrivalStatus      *string `json:"arrival_status,omitempty"`
	ArrivalDate        *string `json:"arrival_date,omitempty"`
	ArrivalCity        *string `json:"arrival_city,omitempty"`
	DestinationCountry *string `json:"destination_country,omitempty"`
	DepartureDate      *string `json:"departure_date,omitempty"`
	Notes              *string `json:"notes,omitempty"`
}

type SelectionProgressHandler struct {
	progressService *service.SelectionProgressService
	documentStorage secureDocumentStorage
	userRepo        domain.UserRepository
}

func NewSelectionProgressHandler(progressService *service.SelectionProgressService, userRepo domain.UserRepository) *SelectionProgressHandler {
	return &SelectionProgressHandler{
		progressService: progressService,
		userRepo:        userRepo,
	}
}

func (h *SelectionProgressHandler) SetDocumentStorage(storage secureDocumentStorage) {
	h.documentStorage = storage
}

// GetProgress retrieves progress for a selection
func (h *SelectionProgressHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	selectionID := chi.URLParam(r, "id")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	progress, err := h.progressService.GetProgress(selectionID)
	if err != nil {
		h.writeProgressError(w, err)
		return
	}

	response := h.mapProgressResponse(progress)
	_ = utils.WriteJSON(w, http.StatusOK, map[string]SelectionProgressResponse{
		"progress": response,
	})
}

// UpdateProgress updates progress fields (Ethiopian agents only)
func (h *SelectionProgressHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	selectionID := chi.URLParam(r, "id")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req UpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Parse arrival date if provided
	var arrivalDate *time.Time
	if req.ArrivalDate != nil && strings.TrimSpace(*req.ArrivalDate) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ArrivalDate))
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid arrival date format"})
			return
		}
		arrivalDate = &parsed
	}

	// Parse departure date if provided
	var departureDate *time.Time
	if req.DepartureDate != nil && strings.TrimSpace(*req.DepartureDate) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.DepartureDate))
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid departure date format"})
			return
		}
		departureDate = &parsed
	}

	input := service.UpdateProgressInput{
		COCStatus:          req.COCStatus,
		COCType:            req.COCType,
		MedicalStatus:      req.MedicalStatus,
		VisaStatus:         req.VisaStatus,
		TicketStatus:       req.TicketStatus,
		ArrivalStatus:      req.ArrivalStatus,
		ArrivalDate:        arrivalDate,
		ArrivalCity:        req.ArrivalCity,
		DestinationCountry: req.DestinationCountry,
		DepartureDate:      departureDate,
		Notes:              req.Notes,
	}

	progress, err := h.progressService.UpdateProgress(selectionID, userID, input)
	if err != nil {
		h.writeProgressError(w, err)
		return
	}

	response := h.mapProgressResponse(progress)
	_ = utils.WriteJSON(w, http.StatusOK, map[string]SelectionProgressResponse{
		"progress": response,
	})
}

// UploadDocument uploads a document for a progress field
func (h *SelectionProgressHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	selectionID := chi.URLParam(r, "id")
	documentType := chi.URLParam(r, "type")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 52<<20)
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			_ = utils.WriteJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "file is too large"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	progress, err := h.progressService.UploadDocument(selectionID, userID, service.UploadProgressDocumentInput{
		DocumentType: documentType,
		File:         file,
		FileName:     fileHeader.Filename,
		FileSize:     fileHeader.Size,
	})
	if err != nil {
		h.writeProgressError(w, err)
		return
	}

	response := h.mapProgressResponse(progress)
	_ = utils.WriteJSON(w, http.StatusOK, map[string]SelectionProgressResponse{
		"progress": response,
	})
}

// DeleteDocument removes a document from a progress field
func (h *SelectionProgressHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	selectionID := chi.URLParam(r, "id")
	documentType := chi.URLParam(r, "type")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	progress, err := h.progressService.DeleteDocument(selectionID, userID, documentType)
	if err != nil {
		h.writeProgressError(w, err)
		return
	}

	response := h.mapProgressResponse(progress)
	_ = utils.WriteJSON(w, http.StatusOK, map[string]SelectionProgressResponse{
		"progress": response,
	})
}

type BatchUpdateProgressRequest struct {
	SelectionIDs       []string `json:"selection_ids"`
	COCStatus          *string  `json:"coc_status,omitempty"`
	COCType            *string  `json:"coc_type,omitempty"`
	MedicalStatus      *string  `json:"medical_status,omitempty"`
	VisaStatus         *string  `json:"visa_status,omitempty"`
	TicketStatus       *string  `json:"ticket_status,omitempty"`
	ArrivalStatus      *string  `json:"arrival_status,omitempty"`
	ArrivalDate        *string  `json:"arrival_date,omitempty"`
	ArrivalCity        *string  `json:"arrival_city,omitempty"`
	DestinationCountry *string  `json:"destination_country,omitempty"`
	DepartureDate      *string  `json:"departure_date,omitempty"`
	Notes              *string  `json:"notes,omitempty"`
}

// BatchUpdateProgress applies the same progress update to multiple selections
func (h *SelectionProgressHandler) BatchUpdateProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req BatchUpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.SelectionIDs) == 0 {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "selection_ids is required"})
		return
	}

	var arrivalDate *time.Time
	if req.ArrivalDate != nil && strings.TrimSpace(*req.ArrivalDate) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ArrivalDate))
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid arrival date format"})
			return
		}
		arrivalDate = &parsed
	}

	var departureDate *time.Time
	if req.DepartureDate != nil && strings.TrimSpace(*req.DepartureDate) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.DepartureDate))
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid departure date format"})
			return
		}
		departureDate = &parsed
	}

	input := service.UpdateProgressInput{
		COCStatus:          req.COCStatus,
		COCType:            req.COCType,
		MedicalStatus:      req.MedicalStatus,
		VisaStatus:         req.VisaStatus,
		TicketStatus:       req.TicketStatus,
		ArrivalStatus:      req.ArrivalStatus,
		ArrivalDate:        arrivalDate,
		ArrivalCity:        req.ArrivalCity,
		DestinationCountry: req.DestinationCountry,
		DepartureDate:      departureDate,
		Notes:              req.Notes,
	}

	results := h.progressService.BatchUpdateProgress(req.SelectionIDs, userID, input)

	updated := 0
	failed := make([]map[string]string, 0)
	for _, result := range results {
		if result.Error != nil {
			failed = append(failed, map[string]string{
				"selection_id": result.SelectionID,
				"error":        result.Error.Error(),
			})
		} else {
			updated++
		}
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"updated": updated,
		"total":   len(req.SelectionIDs),
		"failed":  failed,
	})
}

func (h *SelectionProgressHandler) writeProgressError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrProgressNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "progress not found"})
	case strings.HasPrefix(err.Error(), "invalid "):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case strings.HasPrefix(err.Error(), "cannot mark "):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case strings.HasPrefix(err.Error(), "add a failure reason"):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrProgressAlreadyExists):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "progress already exists"})
	case errors.Is(err, service.ErrProgressUpdateForbidden):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "only ethiopian agents can update progress"})
	case errors.Is(err, service.ErrInvalidProgressDocument):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid document type"})
	case errors.Is(err, service.ErrFileTooLarge):
		_ = utils.WriteJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "file too large"})
	case errors.Is(err, service.ErrInvalidFileType):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid file type"})
	case errors.Is(err, repository.ErrSelectionNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "selection not found"})
	case errors.Is(err, repository.ErrCandidateNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
	case errors.Is(err, repository.ErrProgressNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "progress not found"})
	default:
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func (h *SelectionProgressHandler) mapProgressResponse(progress *domain.SelectionProgress) SelectionProgressResponse {
	userName := ""
	if h.userRepo != nil && strings.TrimSpace(progress.UpdatedBy) != "" {
		if user, err := h.userRepo.GetByID(progress.UpdatedBy); err == nil && user != nil {
			userName = user.FullName
		}
	}

	response := SelectionProgressResponse{
		ID:          progress.ID,
		SelectionID: progress.SelectionID,
		UpdatedBy:   progress.UpdatedBy,

		COCStatus:     progress.COCStatus,
		MedicalStatus: progress.MedicalStatus,
		VisaStatus:    progress.VisaStatus,
		TicketStatus:  progress.TicketStatus,
		ArrivalStatus: progress.ArrivalStatus,

		ProgressPercentage: computeProgressPercentage(progress),
		UpdatedByName:      userName,

		CreatedAt: progress.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: progress.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// COC Type
	if progress.COCType != nil && strings.TrimSpace(*progress.COCType) != "" {
		response.COCType = progress.COCType
	}

	// COC Document
	if strings.TrimSpace(progress.COCDocumentURL) != "" {
		response.COCDocument = h.mapDocumentSummary(
			progress.COCDocumentURL,
			progress.COCDocumentFileName,
			progress.COCDocumentUploadedAt,
		)
	}

	// Medical Document
	if strings.TrimSpace(progress.MedicalDocumentURL) != "" {
		response.MedicalDocument = h.mapDocumentSummary(
			progress.MedicalDocumentURL,
			progress.MedicalDocumentFileName,
			progress.MedicalDocumentUploadedAt,
		)
	}

	// Visa Document
	if strings.TrimSpace(progress.VisaDocumentURL) != "" {
		response.VisaDocument = h.mapDocumentSummary(
			progress.VisaDocumentURL,
			progress.VisaDocumentFileName,
			progress.VisaDocumentUploadedAt,
		)
	}

	// Ticket Document
	if strings.TrimSpace(progress.TicketDocumentURL) != "" {
		response.TicketDocument = h.mapDocumentSummary(
			progress.TicketDocumentURL,
			progress.TicketDocumentFileName,
			progress.TicketDocumentUploadedAt,
		)
	}

	// Arrival Date
	if progress.ArrivalDate != nil {
		dateStr := progress.ArrivalDate.UTC().Format(time.RFC3339)
		response.ArrivalDate = &dateStr
	}

	// Arrival City
	if progress.ArrivalCity != nil && strings.TrimSpace(*progress.ArrivalCity) != "" {
		response.ArrivalCity = progress.ArrivalCity
	}

	// Destination Country
	if progress.DestinationCountry != nil && strings.TrimSpace(*progress.DestinationCountry) != "" {
		response.DestinationCountry = progress.DestinationCountry
	}

	// Departure Date
	if progress.DepartureDate != nil {
		dateStr := progress.DepartureDate.UTC().Format(time.RFC3339)
		response.DepartureDate = &dateStr
	}

	// Notes
	response.Notes = progress.Notes

	// Arrival Document
	if strings.TrimSpace(progress.ArrivalDocumentURL) != "" {
		response.ArrivalDocument = h.mapDocumentSummary(
			progress.ArrivalDocumentURL,
			progress.ArrivalDocumentFileName,
			progress.ArrivalDocumentUploadedAt,
		)
	}

	return response
}

func (h *SelectionProgressHandler) mapDocumentSummary(fileURL, fileName string, uploadedAt *time.Time) *SelectionDocumentSummary {
	if strings.TrimSpace(fileURL) == "" {
		return nil
	}

	summary := &SelectionDocumentSummary{
		FileURL:  buildSignedDocumentURL(h.documentStorage, fileURL, fileName, contentTypeFromFileName(fileName), true),
		FileName: fileName,
	}
	if uploadedAt != nil {
		summary.UploadedAt = uploadedAt.UTC().Format(time.RFC3339)
	}
	return summary
}
