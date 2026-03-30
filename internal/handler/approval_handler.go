package handler

import (
	"errors"
	"io"
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

type RejectSelectionRequest struct {
	Reason string `json:"reason"`
}

type ApprovalRecordResponse struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	Role      string `json:"role"`
	Decision  string `json:"decision"`
	DecidedAt string `json:"decided_at"`
}

type ApprovalStatusResponse struct {
	SelectionID         string                   `json:"selection_id"`
	Status              string                   `json:"status"`
	Approvals           []ApprovalRecordResponse `json:"approvals"`
	IsFullyApproved     bool                     `json:"is_fully_approved"`
	PendingApprovalFrom []string                 `json:"pending_approval_from"`
}

type ApprovalHandler struct {
	approvalService  *service.ApprovalService
	selectionService *service.SelectionService
	candidateRepo    domain.CandidateRepository
	pairingService   *service.PairingService
}

func NewApprovalHandler(approvalService *service.ApprovalService, selectionService *service.SelectionService, candidateRepo domain.CandidateRepository, pairingService *service.PairingService) *ApprovalHandler {
	return &ApprovalHandler{
		approvalService:  approvalService,
		selectionService: selectionService,
		candidateRepo:    candidateRepo,
		pairingService:   pairingService,
	}
}

func (h *ApprovalHandler) ApproveSelection(w http.ResponseWriter, r *http.Request) {
	selectionID := chi.URLParam(r, "id")
	userID, role, ok := authContext(r)
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	pairingID, _ := middleware.PairingIDFromContext(r.Context())

	if err := h.approvalService.ApproveSelection(selectionID, userID); err != nil {
		h.writeApprovalError(w, err)
		return
	}

	response, statusCode, err := h.buildApprovalStatusResponse(selectionID, role, userID, pairingID)
	if err != nil {
		h.writeApprovalError(w, err)
		return
	}

	_ = utils.WriteJSON(w, statusCode, response)
}

func (h *ApprovalHandler) RejectSelection(w http.ResponseWriter, r *http.Request) {
	selectionID := chi.URLParam(r, "id")
	userID, _, ok := authContext(r)
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req RejectSelectionRequest
	if err := decodeJSONBody(w, r, &req, 8<<10); err != nil && !errors.Is(err, io.EOF) {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.approvalService.RejectSelection(selectionID, userID, req.Reason); err != nil {
		h.writeApprovalError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "selection rejected",
	})
}

func (h *ApprovalHandler) GetApprovals(w http.ResponseWriter, r *http.Request) {
	selectionID := chi.URLParam(r, "id")
	userID, role, ok := authContext(r)
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	pairingID, _ := middleware.PairingIDFromContext(r.Context())

	response, statusCode, err := h.buildApprovalStatusResponse(selectionID, role, userID, pairingID)
	if err != nil {
		h.writeApprovalError(w, err)
		return
	}

	_ = utils.WriteJSON(w, statusCode, response)
}

func (h *ApprovalHandler) buildApprovalStatusResponse(selectionID, role, userID, pairingID string) (*ApprovalStatusResponse, int, error) {
	selection, err := h.selectionService.GetSelection(selectionID)
	if err != nil {
		return nil, 0, err
	}
	if h.pairingService != nil {
		pairing, err := h.pairingService.ResolveActivePairing(userID, role, pairingID)
		if err != nil {
			return nil, 0, err
		}
		if strings.TrimSpace(selection.PairingID) != strings.TrimSpace(pairing.ID) {
			return nil, 0, service.ErrNotAuthorized
		}
	}

	candidate, err := h.candidateRepo.GetByID(selection.CandidateID)
	if err != nil {
		return nil, 0, err
	}

	if !isInvolvedSelectionParty(role, userID, selection, candidate) {
		return nil, 0, service.ErrNotAuthorized
	}

	approvals, err := h.approvalService.GetApprovals(selectionID)
	if err != nil {
		return nil, 0, err
	}

	approvalResponses := make([]ApprovalRecordResponse, 0, len(approvals))
	ownerApproved := false
	selectorApproved := false

	for _, approval := range approvals {
		if approval == nil {
			continue
		}
		userName := ""
		userRole := ""
		if approval.User != nil {
			userName = approval.User.FullName
			userRole = string(approval.User.Role)
		}

		approvalResponses = append(approvalResponses, ApprovalRecordResponse{
			UserID:    approval.UserID,
			UserName:  userName,
			Role:      userRole,
			Decision:  string(approval.Decision),
			DecidedAt: approval.DecidedAt.UTC().Format(time.RFC3339),
		})

		if approval.Decision == domain.ApprovalApproved && strings.TrimSpace(approval.UserID) == strings.TrimSpace(candidate.CreatedBy) {
			ownerApproved = true
		}
		if approval.Decision == domain.ApprovalApproved && strings.TrimSpace(approval.UserID) == strings.TrimSpace(selection.SelectedBy) {
			selectorApproved = true
		}
	}

	pending := make([]string, 0, 2)
	if !ownerApproved {
		pending = append(pending, string(domain.EthiopianAgent))
	}
	if !selectorApproved {
		pending = append(pending, string(domain.ForeignAgent))
	}

	response := &ApprovalStatusResponse{
		SelectionID:         selection.ID,
		Status:              string(selection.Status),
		Approvals:           approvalResponses,
		IsFullyApproved:     ownerApproved && selectorApproved && selection.Status == domain.SelectionApproved,
		PendingApprovalFrom: pending,
	}

	return response, http.StatusOK, nil
}

func (h *ApprovalHandler) writeApprovalError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotAuthorized):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, service.ErrSelectionSupportingDocumentsRequired):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrPairingRequired):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "select a partner workspace to continue"})
	case errors.Is(err, service.ErrNoActivePairings), errors.Is(err, service.ErrPairingAccessDenied):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, service.ErrSelectionNotPending), errors.Is(err, service.ErrAlreadyDecided):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, repository.ErrSelectionNotFound), errors.Is(err, repository.ErrCandidateNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	default:
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func authContext(r *http.Request) (string, string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		return "", "", false
	}
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok || strings.TrimSpace(role) == "" {
		return "", "", false
	}
	return strings.TrimSpace(userID), strings.TrimSpace(role), true
}

func isInvolvedSelectionParty(role, userID string, selection *domain.Selection, candidate *domain.Candidate) bool {
	switch strings.TrimSpace(role) {
	case string(domain.ForeignAgent):
		return strings.TrimSpace(selection.SelectedBy) == strings.TrimSpace(userID)
	case string(domain.EthiopianAgent):
		return strings.TrimSpace(candidate.CreatedBy) == strings.TrimSpace(userID)
	default:
		return false
	}
}
