package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
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
	selectionService        *service.SelectionService
	candidateRepo           domain.CandidateRepository
	approvalRepo            domain.ApprovalRepository
	pairingService          *service.PairingService
	documentStorage         secureDocumentStorage
	selectionUpdatesHandler *SelectionUpdatesHandler
}

func NewSelectionHandler(selectionService *service.SelectionService, candidateRepo domain.CandidateRepository, approvalRepo domain.ApprovalRepository, pairingService *service.PairingService) *SelectionHandler {
	return &SelectionHandler{selectionService: selectionService, candidateRepo: candidateRepo, approvalRepo: approvalRepo, pairingService: pairingService}
}

func (h *SelectionHandler) SetDocumentStorage(storage secureDocumentStorage) {
	h.documentStorage = storage
}

func (h *SelectionHandler) SetSelectionUpdatesHandler(handler *SelectionUpdatesHandler) {
	h.selectionUpdatesHandler = handler
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

	// Candidate data is loaded in selection service, just fetch it fresh for the response
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

	// Candidate data is already preloaded by the repository
	if !isInvolvedParty(role, userID, selection, selection.Candidate) {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	response, err := h.mapSelectionResponse(selection, selection.Candidate)
	if err != nil {
		h.writeSelectionError(w, err)
		return
	}

	responseData := map[string]SelectionResponse{
		"selection": response,
	}

	// Add caching headers
	w.Header().Set("Cache-Control", "public, max-age=30")

	// Calculate and set ETag for response caching
	etag := h.calculateETag(responseData)
	w.Header().Set("ETag", etag)

	_ = utils.WriteJSON(w, http.StatusOK, responseData)
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

	// Support server-side sorting
	sortBy := r.URL.Query().Get("sortBy")
	if sortBy == "expiring" {
		// Sort by expiring soon for pending selections
		h.sortSelectionsByExpiring(selections)
	}
	// Default is "newest" (already sorted by created_at DESC in repository)

	// Apply pagination
	limit := 25
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Store total count before pagination
	totalCount := len(selections)

	// Apply pagination to selections
	endIdx := offset + limit
	if endIdx > len(selections) {
		endIdx = len(selections)
	}
	if offset >= len(selections) {
		offset = len(selections)
		endIdx = len(selections)
	}
	paginatedSelections := selections[offset:endIdx]

	responses := make([]SelectionResponse, 0, len(paginatedSelections))
	for _, selection := range paginatedSelections {
		// Candidate data is already preloaded by the repository
		response, err := h.mapSelectionResponse(selection, selection.Candidate)
		if err != nil {
			h.writeSelectionError(w, err)
			return
		}
		responses = append(responses, response)
	}

	// Create response with pagination info
	responseData := map[string]interface{}{
		"selections": responses,
		"pagination": map[string]interface{}{
			"limit":    limit,
			"offset":   offset,
			"total":    totalCount,
			"has_more": endIdx < totalCount,
		},
	}

	// Add caching headers
	w.Header().Set("Cache-Control", "public, max-age=30")

	// Calculate and set ETag for response caching
	etag := h.calculateETag(responseData)
	w.Header().Set("ETag", etag)

	_ = utils.WriteJSON(w, http.StatusOK, responseData)
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

// sortSelectionsByExpiring sorts selections by expiring soon
// Pending selections are listed first (sorted by expires_at), then other selections (sorted by created_at)
func (h *SelectionHandler) sortSelectionsByExpiring(selections []*domain.Selection) {
	sort.SliceStable(selections, func(i, j int) bool {
		iPending := selections[i].Status == domain.SelectionPending
		jPending := selections[j].Status == domain.SelectionPending

		// Pending selections come first
		if iPending && !jPending {
			return true
		}
		if !iPending && jPending {
			return false
		}

		// Both pending: sort by expires_at
		if iPending && jPending {
			return selections[i].ExpiresAt.Before(selections[j].ExpiresAt)
		}

		// Both non-pending: sort by created_at (newest first)
		return selections[i].CreatedAt.After(selections[j].CreatedAt)
	})
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
			photoURL = buildSignedDocumentURL(h.documentStorage, document.FileURL, document.FileName, contentTypeFromFileName(document.FileName), true)
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
		EmployerContract:  h.mapSelectionDocumentSummary(selection.EmployerContractURL, selection.EmployerContractFileName, selection.EmployerContractUploadedAt),
		EmployerID:        h.mapSelectionDocumentSummary(selection.EmployerIDURL, selection.EmployerIDFileName, selection.EmployerIDUploadedAt),
		CreatedAt:         selection.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         selection.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (h *SelectionHandler) mapSelectionDocumentSummary(fileURL, fileName string, uploadedAt *time.Time) *SelectionDocumentSummary {
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

// calculateETag computes an ETag hash for response caching
func (h *SelectionHandler) calculateETag(data interface{}) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		// Fallback to current timestamp if marshaling fails
		return fmt.Sprintf(`"%d"`, time.Now().Unix())
	}

	hash := md5.Sum(jsonBytes)
	return fmt.Sprintf(`"%s"`, hex.EncodeToString(hash[:]))
}
