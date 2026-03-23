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

type PairingAgencySummary struct {
	ID          string `json:"id"`
	FullName    string `json:"full_name"`
	CompanyName string `json:"company_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
}

type PairingWorkspaceSummary struct {
	ID               string               `json:"id"`
	Status           string               `json:"status"`
	EthiopianAgency  PairingAgencySummary `json:"ethiopian_agency"`
	ForeignAgency    PairingAgencySummary `json:"foreign_agency"`
	PartnerAgency    PairingAgencySummary `json:"partner_agency"`
	ApprovedAt       string               `json:"approved_at,omitempty"`
	Notes            string               `json:"notes,omitempty"`
}

type PairingContextResponse struct {
	CurrentUserRole string                   `json:"current_user_role"`
	HasActivePairs  bool                     `json:"has_active_pairs"`
	ActivePairingID string                   `json:"active_pairing_id,omitempty"`
	Workspaces      []PairingWorkspaceSummary `json:"workspaces"`
}

type CandidatePairShareResponse struct {
	ID           string                 `json:"id"`
	PairingID    string                 `json:"pairing_id"`
	SharedAt     string                 `json:"shared_at"`
	IsActive     bool                   `json:"is_active"`
	PartnerAgency PairingAgencySummary  `json:"partner_agency"`
	Workspace    PairingWorkspaceSummary `json:"workspace"`
}

type PairingHandler struct {
	pairingService     *service.PairingService
	userRepository     domain.UserRepository
	candidateRepository *repository.GormCandidateRepository
}

func NewPairingHandler(pairingService *service.PairingService, userRepository domain.UserRepository, candidateRepository *repository.GormCandidateRepository) *PairingHandler {
	return &PairingHandler{
		pairingService:      pairingService,
		userRepository:      userRepository,
		candidateRepository: candidateRepository,
	}
}

func (h *PairingHandler) GetMyPairingContext(w http.ResponseWriter, r *http.Request) {
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

	pairings, err := h.pairingService.ListActivePairingsForUser(userID)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	pairingID, _ := middleware.PairingIDFromContext(r.Context())
	response := PairingContextResponse{
		CurrentUserRole: strings.TrimSpace(role),
		HasActivePairs:  len(pairings) > 0,
		Workspaces:      make([]PairingWorkspaceSummary, 0, len(pairings)),
	}

	for _, pairing := range pairings {
		workspace, err := h.mapWorkspace(pairing, userID, role)
		if err != nil {
			continue
		}
		response.Workspaces = append(response.Workspaces, workspace)
		if response.ActivePairingID == "" && (pairingID == "" || pairing.ID == pairingID) {
			response.ActivePairingID = pairing.ID
		}
	}

	if response.ActivePairingID == "" && len(response.Workspaces) == 1 {
		response.ActivePairingID = response.Workspaces[0].ID
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]PairingContextResponse{"context": response})
}

func (h *PairingHandler) ShareCandidate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	pairingID := chi.URLParam(r, "pairingId")
	candidateID := chi.URLParam(r, "candidateId")

	candidate, err := h.candidateRepository.GetByID(candidateID)
	if err != nil {
		if errors.Is(err, repository.ErrCandidateNotFound) {
			_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(userID) {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	if candidate.Status != domain.CandidateStatusDraft && candidate.Status != domain.CandidateStatusAvailable {
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "candidate cannot be shared in its current status"})
		return
	}

	if err := h.pairingService.ShareCandidate(candidateID, pairingID, userID); err != nil {
		h.writePairingError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "candidate shared"})
}

func (h *PairingHandler) UnshareCandidate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	pairingID := chi.URLParam(r, "pairingId")
	candidateID := chi.URLParam(r, "candidateId")
	if err := h.pairingService.UnshareCandidate(candidateID, pairingID, userID); err != nil {
		h.writePairingError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "candidate unshared"})
}

func (h *PairingHandler) GetCandidateShares(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	candidateID := chi.URLParam(r, "id")
	candidate, err := h.candidateRepository.GetByID(candidateID)
	if err != nil {
		if errors.Is(err, repository.ErrCandidateNotFound) {
			_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(userID) {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	shares, err := h.pairingService.ListCandidateShares(candidateID)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	response := make([]CandidatePairShareResponse, 0, len(shares))
	for _, share := range shares {
		pairing, err := h.pairingService.ResolveActivePairing(userID, string(domain.EthiopianAgent), share.PairingID)
		if err != nil {
			if errors.Is(err, service.ErrPairingAccessDenied) || errors.Is(err, service.ErrPairingNotFound) || errors.Is(err, service.ErrNoActivePairings) {
				continue
			}
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		workspace, err := h.mapWorkspace(pairing, userID, string(domain.EthiopianAgent))
		if err != nil {
			continue
		}
		response = append(response, CandidatePairShareResponse{
			ID:            share.ID,
			PairingID:     share.PairingID,
			SharedAt:      share.SharedAt.UTC().Format(time.RFC3339),
			IsActive:      share.IsActive,
			PartnerAgency: workspace.PartnerAgency,
			Workspace:     workspace,
		})
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string][]CandidatePairShareResponse{"shares": response})
}

func (h *PairingHandler) mapWorkspace(pairing *domain.AgencyPairing, userID, role string) (PairingWorkspaceSummary, error) {
	ethiopianUser, err := h.userRepository.GetByID(pairing.EthiopianUserID)
	if err != nil {
		return PairingWorkspaceSummary{}, err
	}
	foreignUser, err := h.userRepository.GetByID(pairing.ForeignUserID)
	if err != nil {
		return PairingWorkspaceSummary{}, err
	}

	ethiopianSummary := mapPairingAgencySummary(ethiopianUser)
	foreignSummary := mapPairingAgencySummary(foreignUser)
	partner := foreignSummary
	if strings.TrimSpace(role) == string(domain.ForeignAgent) {
		partner = ethiopianSummary
	}

	workspace := PairingWorkspaceSummary{
		ID:              pairing.ID,
		Status:          string(pairing.Status),
		EthiopianAgency: ethiopianSummary,
		ForeignAgency:   foreignSummary,
		PartnerAgency:   partner,
	}
	if pairing.ApprovedAt != nil {
		workspace.ApprovedAt = pairing.ApprovedAt.UTC().Format(time.RFC3339)
	}
	if pairing.Notes != nil {
		workspace.Notes = *pairing.Notes
	}

	return workspace, nil
}

func mapPairingAgencySummary(user *domain.User) PairingAgencySummary {
	if user == nil {
		return PairingAgencySummary{}
	}
	return PairingAgencySummary{
		ID:          user.ID,
		FullName:    user.FullName,
		CompanyName: user.CompanyName,
		Email:       user.Email,
		Role:        string(user.Role),
	}
}

func (h *PairingHandler) writePairingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrPairingRequired):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "pairing is required"})
	case errors.Is(err, service.ErrNoActivePairings):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "no active pairings"})
	case errors.Is(err, service.ErrPairingAccessDenied):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, service.ErrPairingNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "pairing not found"})
	case errors.Is(err, service.ErrPairingNotActive):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "pairing is not active"})
	case errors.Is(err, service.ErrCandidateAlreadyShared):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "candidate already shared to this workspace"})
	case errors.Is(err, service.ErrCandidateShareNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "candidate share not found"})
	default:
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
