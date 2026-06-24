package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type UpdatedByResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type StatusStepResponse struct {
	ID                 string            `json:"id"`
	CandidateID        string            `json:"candidate_id"`
	StepName           string            `json:"step_name"`
	StepStatus         string            `json:"step_status"`
	CompletedAt        *string           `json:"completed_at,omitempty"`
	Notes              string            `json:"notes,omitempty"`
	MedicalDocumentURL *string           `json:"medical_document_url,omitempty"`
	CoCStatus          *string           `json:"coc_status,omitempty"`
	ArrivalCity        *string           `json:"arrival_city,omitempty"`
	UpdatedBy          UpdatedByResponse `json:"updated_by"`
	UpdatedAt          string            `json:"updated_at"`
}

type ProgressResponse struct {
	CandidateID        string               `json:"candidate_id"`
	OverallStatus      string               `json:"overall_status"`
	Steps              []StatusStepResponse `json:"steps"`
	ProgressPercentage float64              `json:"progress_percentage"`
	LastUpdatedAt      string               `json:"last_updated_at"`
}

type UpdateStatusStepRequest struct {
	Status      string  `json:"status"`
	Notes       string  `json:"notes"`
	CoCStatus   *string `json:"coc_status,omitempty"`
	ArrivalCity *string `json:"arrival_city,omitempty"`
}

type BatchUpdateStatusRequest struct {
	CandidateIDs []string `json:"candidate_ids"`
	StepName     string   `json:"step_name"`
	Status       string   `json:"status"`
	Notes        string   `json:"notes,omitempty"`
	CoCStatus    *string  `json:"coc_status,omitempty"`
	ArrivalCity  *string  `json:"arrival_city,omitempty"`
}

type StatusHandler struct {
	statusStepService *service.StatusStepService
	candidateRepo     domain.CandidateRepository
	selectionRepo     domain.SelectionRepository
	documentRepo      domain.DocumentRepository
	userRepo          domain.UserRepository
	pairingService    *service.PairingService
}

func NewStatusHandler(statusStepService *service.StatusStepService, candidateRepo domain.CandidateRepository, selectionRepo domain.SelectionRepository, documentRepo domain.DocumentRepository, userRepo domain.UserRepository, pairingService *service.PairingService) *StatusHandler {
	return &StatusHandler{
		statusStepService: statusStepService,
		candidateRepo:     candidateRepo,
		selectionRepo:     selectionRepo,
		documentRepo:      documentRepo,
		userRepo:          userRepo,
		pairingService:    pairingService,
	}
}

func (h *StatusHandler) GetCandidateStatusSteps(w http.ResponseWriter, r *http.Request) {
	candidateID := strings.TrimSpace(chi.URLParam(r, "id"))
	userID, role, ok := authContext(r)
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	pairingID, _ := middleware.PairingIDFromContext(r.Context())

	candidate, err := h.candidateRepo.GetByID(candidateID)
	if err != nil {
		h.writeStatusError(w, err)
		return
	}

	involved, err := h.isInvolved(candidate, userID, role, pairingID)
	if err != nil {
		h.writeStatusError(w, err)
		return
	}
	if !involved {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	response, err := h.buildProgressResponse(candidate)
	if err != nil {
		h.writeStatusError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, response)
}

func (h *StatusHandler) UpdateStatusStep(w http.ResponseWriter, r *http.Request) {
	candidateID := strings.TrimSpace(chi.URLParam(r, "id"))
	stepNameRaw := strings.TrimSpace(chi.URLParam(r, "step_name"))
	stepName, err := url.PathUnescape(stepNameRaw)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid step name"})
		return
	}
	stepName, err = canonicalStepName(strings.TrimSpace(stepName))
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid step name"})
		return
	}

	userID, role, ok := authContext(r)
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if role != string(domain.EthiopianAgent) {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	candidate, err := h.candidateRepo.GetByID(candidateID)
	if err != nil {
		h.writeStatusError(w, err)
		return
	}
	if strings.TrimSpace(candidate.CreatedBy) != userID {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	var req UpdateStatusStepRequest
	if err := decodeJSONBody(w, r, &req, 16<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	status := domain.StepStatus(strings.TrimSpace(req.Status))
	if status != domain.Pending && status != domain.InProgress && status != domain.Completed && status != domain.Failed {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "status must be pending, in_progress, completed, or failed"})
		return
	}

	if err := h.statusStepService.UpdateStepWithExtras(candidateID, stepName, userID, status, req.Notes, service.StepExtras{
		CoCStatus:   req.CoCStatus,
		ArrivalCity: req.ArrivalCity,
	}); err != nil {
		h.writeStatusError(w, err)
		return
	}

	response, err := h.buildProgressResponse(candidate)
	if err != nil {
		h.writeStatusError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, response)
}

func (h *StatusHandler) BatchUpdateStatusSteps(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := authContext(r)
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if role != string(domain.EthiopianAgent) {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	var req BatchUpdateStatusRequest
	if err := decodeJSONBody(w, r, &req, 32<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.CandidateIDs) == 0 {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "candidate_ids is required"})
		return
	}

	stepName, err := canonicalStepName(strings.TrimSpace(req.StepName))
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid step name"})
		return
	}

	status := domain.StepStatus(strings.TrimSpace(req.Status))
	if status != domain.Pending && status != domain.InProgress && status != domain.Completed && status != domain.Failed {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "status must be pending, in_progress, completed, or failed"})
		return
	}

	results := h.statusStepService.BatchUpdateSteps(req.CandidateIDs, stepName, status, req.Notes, req.CoCStatus, req.ArrivalCity, userID)

	updated := 0
	failed := make([]map[string]string, 0)
	for _, result := range results {
		if result.Error != nil {
			failed = append(failed, map[string]string{
				"candidate_id": result.CandidateID,
				"error":        result.Error.Error(),
			})
		} else {
			updated++
		}
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"updated": updated,
		"total":   len(req.CandidateIDs),
		"failed":  failed,
	})
}

func (h *StatusHandler) buildProgressResponse(candidate *domain.Candidate) (*ProgressResponse, error) {
	steps, err := h.statusStepService.GetCandidateProgress(candidate.ID)
	if err != nil {
		return nil, err
	}
	medicalDocumentURL, err := h.resolveMedicalDocumentURL(candidate.ID)
	if err != nil {
		return nil, err
	}

	responses := make([]StatusStepResponse, 0, len(steps))
	completedCount := 0
	lastUpdatedAt := candidate.UpdatedAt.UTC()
	userNameCache := make(map[string]string)

	for _, step := range steps {
		if step == nil {
			continue
		}

		if step.StepStatus == domain.Completed {
			completedCount++
		}
		if step.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = step.UpdatedAt
		}

		userName := userNameCache[step.UpdatedBy]
		if userName == "" && strings.TrimSpace(step.UpdatedBy) != "" {
			user, err := h.userRepo.GetByID(step.UpdatedBy)
			if err == nil && user != nil {
				userName = user.FullName
				userNameCache[step.UpdatedBy] = userName
			}
		}

		var completedAt *string
		if step.CompletedAt != nil {
			formatted := step.CompletedAt.UTC().Format(time.RFC3339)
			completedAt = &formatted
		}
		var stepMedicalDocumentURL *string
		if strings.EqualFold(strings.TrimSpace(step.StepName), strings.TrimSpace(domain.Medical)) && medicalDocumentURL != nil {
			stepMedicalDocumentURL = medicalDocumentURL
		}

		responses = append(responses, StatusStepResponse{
			ID:                 step.ID,
			CandidateID:        step.CandidateID,
			StepName:           step.StepName,
			StepStatus:         string(step.StepStatus),
			CompletedAt:        completedAt,
			Notes:              step.Notes,
			MedicalDocumentURL: stepMedicalDocumentURL,
			CoCStatus:          step.CoCStatus,
			ArrivalCity:        step.ArrivalCity,
			UpdatedBy: UpdatedByResponse{
				ID:   step.UpdatedBy,
				Name: userName,
			},
			UpdatedAt: step.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	progressPercentage := 0.0
	if len(steps) > 0 {
		progressPercentage = (float64(completedCount) / float64(len(steps))) * 100
	}

	return &ProgressResponse{
		CandidateID:        candidate.ID,
		OverallStatus:      string(candidate.Status),
		Steps:              responses,
		ProgressPercentage: progressPercentage,
		LastUpdatedAt:      lastUpdatedAt.Format(time.RFC3339),
	}, nil
}

func (h *StatusHandler) resolveMedicalDocumentURL(candidateID string) (*string, error) {
	if h.documentRepo == nil {
		return nil, nil
	}

	documents, err := h.documentRepo.GetByCandidateID(candidateID)
	if err != nil {
		return nil, err
	}

	for _, document := range documents {
		if document == nil || document.DocumentType != domain.MedicalDocument {
			continue
		}
		if strings.TrimSpace(document.FileURL) == "" {
			continue
		}
		value := document.FileURL
		return &value, nil
	}

	return nil, nil
}

func (h *StatusHandler) isInvolved(candidate *domain.Candidate, userID, role, pairingID string) (bool, error) {
	if strings.TrimSpace(role) == string(domain.EthiopianAgent) {
		return strings.TrimSpace(candidate.CreatedBy) == strings.TrimSpace(userID), nil
	}
	if strings.TrimSpace(role) == string(domain.ForeignAgent) {
		if h.pairingService != nil {
			pairing, err := h.pairingService.ResolveActivePairing(userID, role, pairingID)
			if err != nil {
				return false, err
			}
			selection, err := h.selectionRepo.GetByCandidateIDAndPairingID(candidate.ID, pairing.ID)
			if err != nil {
				if errors.Is(err, repository.ErrSelectionNotFound) {
					return false, nil
				}
				return false, err
			}
			return strings.TrimSpace(selection.SelectedBy) == strings.TrimSpace(userID), nil
		}
		selection, err := h.selectionRepo.GetByCandidateID(candidate.ID)
		if err != nil {
			return false, nil
		}
		return strings.TrimSpace(selection.SelectedBy) == strings.TrimSpace(userID), nil
	}
	return false, nil
}

func (h *StatusHandler) writeStatusError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrCandidateNotFound), errors.Is(err, repository.ErrSelectionNotFound), errors.Is(err, service.ErrStepNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, service.ErrNotAuthorized):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, service.ErrPairingRequired):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "select a partner workspace to continue"})
	case errors.Is(err, service.ErrNoActivePairings), errors.Is(err, service.ErrPairingAccessDenied):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, service.ErrInvalidStepTransition), errors.Is(err, service.ErrStepFailureReasonRequired), errors.Is(err, service.ErrMedicalDocumentRequired):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	default:
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func normalizeStepName(stepName string) string {
	return strings.ToLower(strings.TrimSpace(stepName))
}

func canonicalStepName(stepName string) (string, error) {
	normalized := normalizeStepName(stepName)
	mapping := map[string]string{
		normalizeStepName(domain.Medical):         domain.Medical,
		normalizeStepName(domain.CoC):             domain.CoC,
		normalizeStepName(domain.Visa):            domain.Visa,
		normalizeStepName(domain.Ticket):          domain.Ticket,
		normalizeStepName(domain.ArrivalCity):     domain.ArrivalCity,

		// Legacy mappings
		normalizeStepName(domain.CoCPending):      domain.CoC,
		normalizeStepName(domain.CoCOnline):       domain.CoC,
		normalizeStepName(domain.LMISPending):     domain.Visa,
		normalizeStepName(domain.LMISIssued):      domain.Visa,
		normalizeStepName(domain.TicketPending):   domain.Ticket,
		normalizeStepName(domain.TicketBooked):    domain.Ticket,
		normalizeStepName(domain.TicketConfirmed): domain.Ticket,
		normalizeStepName(domain.Arrived):         domain.ArrivalCity,
	}
	value, ok := mapping[normalized]
	if !ok {
		return "", fmt.Errorf("invalid step name")
	}
	return value, nil
}
