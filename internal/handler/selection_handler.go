package handler

import (
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

type SelectionCandidateSummary struct {
	ID              string `json:"id"`
	FullName        string `json:"full_name"`
	Status          string `json:"status"`
	CreatedBy       string `json:"created_by"`
	Age             *int   `json:"age,omitempty"`
	ExperienceYears *int   `json:"experience_years,omitempty"`
	PhotoURL        string `json:"photo_url,omitempty"`
}

type SelectionResponse struct {
	ID                string                    `json:"id"`
	CandidateID       string                    `json:"candidate_id"`
	SelectedBy        string                    `json:"selected_by"`
	Status            string                    `json:"status"`
	ExpiresAt         string                    `json:"expires_at"`
	TimeRemaining     string                    `json:"time_remaining"`
	Candidate         SelectionCandidateSummary `json:"candidate"`
	EthiopianApproved bool                      `json:"ethiopian_approved"`
	ForeignApproved   bool                      `json:"foreign_approved"`
	EmployerContract  *SelectionDocumentSummary `json:"employer_contract,omitempty"`
	EmployerID        *SelectionDocumentSummary `json:"employer_id,omitempty"`
	CreatedAt         string                    `json:"created_at"`
	UpdatedAt         string                    `json:"updated_at"`
}

type SelectionDocumentSummary struct {
	FileURL    string `json:"file_url"`
	FileName   string `json:"file_name"`
	UploadedAt string `json:"uploaded_at"`
}

type SelectionHandler struct {
	selectionService *service.SelectionService
	candidateRepo    domain.CandidateRepository
	approvalRepo     domain.ApprovalRepository
	pairingService   *service.PairingService
}

func NewSelectionHandler(selectionService *service.SelectionService, candidateRepo domain.CandidateRepository, approvalRepo domain.ApprovalRepository, pairingService *service.PairingService) *SelectionHandler {
	return &SelectionHandler{selectionService: selectionService, candidateRepo: candidateRepo, approvalRepo: approvalRepo, pairingService: pairingService}
}

func (h *SelectionHandler) SelectCandidate(w http.ResponseWriter, r *http.Request) {
	candidateID := chi.URLParam(r, "id")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	pairingID, _ := middleware.PairingIDFromContext(r.Context())

	selection, err := h.selectionService.SelectCandidateInPairing(candidateID, userID, pairingID)
	if err != nil {
		if h.tryWritePairingError(w, err) {
			return
		}
		h.writeSelectionError(w, err)
		return
	}

	candidate, err := h.candidateRepo.GetByID(selection.CandidateID)
	if err != nil {
		h.writeSelectionError(w, err)
		return
	}

	response, err := h.mapSelectionResponse(selection, candidate)
	if err != nil {
		h.writeSelectionError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusCreated, map[string]SelectionResponse{
		"selection": response,
	})
}

func (h *SelectionHandler) GetSelection(w http.ResponseWriter, r *http.Request) {
	selectionID := chi.URLParam(r, "id")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	role, ok := middleware.RoleFromContext(r.Context())
	if !ok || strings.TrimSpace(role) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	pairingID, _ := middleware.PairingIDFromContext(r.Context())

	selection, err := h.selectionService.GetSelection(selectionID)
	if err != nil {
		h.writeSelectionError(w, err)
		return
	}
	if h.pairingService != nil {
		pairing, err := h.pairingService.ResolveActivePairing(userID, role, pairingID)
		if err != nil {
			if h.tryWritePairingError(w, err) {
				return
			}
			h.writeSelectionError(w, err)
			return
		}
		if strings.TrimSpace(selection.PairingID) != strings.TrimSpace(pairing.ID) {
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
	}

	candidate, err := h.candidateRepo.GetByID(selection.CandidateID)
	if err != nil {
		h.writeSelectionError(w, err)
		return
	}

	if !isInvolvedParty(role, userID, selection, candidate) {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	response, err := h.mapSelectionResponse(selection, candidate)
	if err != nil {
		h.writeSelectionError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]SelectionResponse{
		"selection": response,
	})
}

func (h *SelectionHandler) GetMySelections(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	role, ok := middleware.RoleFromContext(r.Context())
	if !ok || strings.TrimSpace(role) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	pairingID, _ := middleware.PairingIDFromContext(r.Context())

	selections, err := h.selectionService.GetSelectionsForWorkspace(userID, role, pairingID)
	if err != nil {
		if h.tryWritePairingError(w, err) {
			return
		}
		h.writeSelectionError(w, err)
		return
	}

	responses := make([]SelectionResponse, 0, len(selections))
	for _, selection := range selections {
		candidate, err := h.candidateRepo.GetByID(selection.CandidateID)
		if err != nil {
			h.writeSelectionError(w, err)
			return
		}
		response, err := h.mapSelectionResponse(selection, candidate)
		if err != nil {
			h.writeSelectionError(w, err)
			return
		}
		responses = append(responses, response)
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string][]SelectionResponse{
		"selections": responses,
	})
}

func (h *SelectionHandler) UploadSelectionDocument(w http.ResponseWriter, r *http.Request) {
	selectionID := chi.URLParam(r, "id")
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

	documentType := strings.TrimSpace(r.FormValue("document_type"))
	if documentType == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "document type is required"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	selection, err := h.selectionService.UploadSelectionDocument(selectionID, userID, service.UploadSelectionDocumentInput{
		DocumentType: documentType,
		File:         file,
		FileName:     fileHeader.Filename,
		FileSize:     fileHeader.Size,
	})
	if err != nil {
		h.writeSelectionError(w, err)
		return
	}

	candidate, err := h.candidateRepo.GetByID(selection.CandidateID)
	if err != nil {
		h.writeSelectionError(w, err)
		return
	}

	response, err := h.mapSelectionResponse(selection, candidate)
	if err != nil {
		h.writeSelectionError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]SelectionResponse{
		"selection": response,
	})
}

func (h *SelectionHandler) writeSelectionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAlreadySelected):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "candidate already selected"})
	case errors.Is(err, service.ErrCandidateNotAvailable):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "candidate not available"})
	case errors.Is(err, service.ErrNotForeignAgent):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "only foreign agents can select candidates"})
	case errors.Is(err, service.ErrNotAuthorized):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, service.ErrSelectionNotPending):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "selection is not pending"})
	case errors.Is(err, service.ErrInvalidSelectionDocumentType), errors.Is(err, service.ErrFileTooLarge), errors.Is(err, service.ErrInvalidFileType):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, repository.ErrSelectionNotFound), errors.Is(err, repository.ErrCandidateNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	default:
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func (h *SelectionHandler) tryWritePairingError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, service.ErrPairingRequired):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "select a partner workspace to continue"})
		return true
	case errors.Is(err, service.ErrNoActivePairings):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "no active partner workspaces"})
		return true
	case errors.Is(err, service.ErrPairingAccessDenied):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return true
	default:
		return false
	}
}

func isInvolvedParty(role, userID string, selection *domain.Selection, candidate *domain.Candidate) bool {
	trimmedRole := strings.TrimSpace(role)
	trimmedUserID := strings.TrimSpace(userID)

	switch trimmedRole {
	case string(domain.ForeignAgent):
		return strings.TrimSpace(selection.SelectedBy) == trimmedUserID
	case string(domain.EthiopianAgent):
		return strings.TrimSpace(candidate.CreatedBy) == trimmedUserID
	default:
		return false
	}
}

func (h *SelectionHandler) mapSelectionResponse(selection *domain.Selection, candidate *domain.Candidate) (SelectionResponse, error) {
	remaining := time.Until(selection.ExpiresAt)
	if remaining < 0 {
		remaining = 0
	}

	approvals, err := h.approvalRepo.GetBySelectionID(selection.ID)
	if err != nil {
		return SelectionResponse{}, err
	}

	ethiopianApproved := false
	foreignApproved := false
	for _, approval := range approvals {
		if approval == nil || approval.Decision != domain.ApprovalApproved {
			continue
		}
		if strings.TrimSpace(approval.UserID) == strings.TrimSpace(candidate.CreatedBy) {
			ethiopianApproved = true
		}
		if strings.TrimSpace(approval.UserID) == strings.TrimSpace(selection.SelectedBy) {
			foreignApproved = true
		}
	}

	photoURL := ""
	for _, document := range candidate.Documents {
		if document.DocumentType == string(domain.Photo) {
			photoURL = document.FileURL
			break
		}
	}

	return SelectionResponse{
		ID:            selection.ID,
		CandidateID:   selection.CandidateID,
		SelectedBy:    selection.SelectedBy,
		Status:        string(selection.Status),
		ExpiresAt:     selection.ExpiresAt.UTC().Format(time.RFC3339),
		TimeRemaining: remaining.Truncate(time.Second).String(),
		Candidate: SelectionCandidateSummary{
			ID:              candidate.ID,
			FullName:        candidate.FullName,
			Status:          string(candidate.Status),
			CreatedBy:       candidate.CreatedBy,
			Age:             candidate.Age,
			ExperienceYears: candidate.ExperienceYears,
			PhotoURL:        photoURL,
		},
		EthiopianApproved: ethiopianApproved,
		ForeignApproved:   foreignApproved,
		EmployerContract:  mapSelectionDocumentSummary(selection.EmployerContractURL, selection.EmployerContractFileName, selection.EmployerContractUploadedAt),
		EmployerID:        mapSelectionDocumentSummary(selection.EmployerIDURL, selection.EmployerIDFileName, selection.EmployerIDUploadedAt),
		CreatedAt:         selection.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         selection.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func mapSelectionDocumentSummary(fileURL, fileName string, uploadedAt *time.Time) *SelectionDocumentSummary {
	if strings.TrimSpace(fileURL) == "" {
		return nil
	}

	summary := &SelectionDocumentSummary{
		FileURL:  fileURL,
		FileName: fileName,
	}
	if uploadedAt != nil {
		summary.UploadedAt = uploadedAt.UTC().Format(time.RFC3339)
	}
	return summary
}
