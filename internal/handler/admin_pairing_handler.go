package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type AdminPairingResponse struct {
	ID               string               `json:"id"`
	Status           string               `json:"status"`
	EthiopianAgency  PairingAgencySummary `json:"ethiopian_agency"`
	ForeignAgency    PairingAgencySummary `json:"foreign_agency"`
	ApprovedAt       string               `json:"approved_at,omitempty"`
	Notes            string               `json:"notes,omitempty"`
}

type AdminCreatePairingRequest struct {
	EthiopianUserID string `json:"ethiopian_user_id"`
	ForeignUserID   string `json:"foreign_user_id"`
	Notes           string `json:"notes"`
}

type AdminUpdatePairingRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

type AdminPairingHandler struct {
	pairingService *service.PairingService
	userRepository domain.UserRepository
}

func NewAdminPairingHandler(pairingService *service.PairingService, userRepository domain.UserRepository) *AdminPairingHandler {
	return &AdminPairingHandler{
		pairingService: pairingService,
		userRepository: userRepository,
	}
}

func (h *AdminPairingHandler) GetPairings(w http.ResponseWriter, r *http.Request) {
	filters := domain.AgencyPairingFilters{
		EthiopianUserID: strings.TrimSpace(r.URL.Query().Get("ethiopian_user_id")),
		ForeignUserID:   strings.TrimSpace(r.URL.Query().Get("foreign_user_id")),
		UserID:          strings.TrimSpace(r.URL.Query().Get("agency_id")),
	}
	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		parsedStatus := domain.AgencyPairingStatus(status)
		filters.Status = &parsedStatus
	}

	pairings, err := h.pairingService.ListAdminPairings(filters)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	items, err := h.mapAdminPairings(pairings)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	_ = utils.WriteJSON(w, http.StatusOK, map[string][]AdminPairingResponse{"pairings": items})
}

func (h *AdminPairingHandler) CreatePairing(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req AdminCreatePairingRequest
	if err := decodeJSONBody(w, r, &req, 32<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	pairing, err := h.pairingService.CreatePairing(adminID, req.EthiopianUserID, req.ForeignUserID, req.Notes, clientIP(r))
	if err != nil {
		h.writeAdminPairingError(w, err)
		return
	}

	items, err := h.mapAdminPairings([]*domain.AgencyPairing{pairing})
	if err != nil || len(items) == 0 {
		_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "pairing created"})
		return
	}
	_ = utils.WriteJSON(w, http.StatusCreated, map[string]AdminPairingResponse{"pairing": items[0]})
}

func (h *AdminPairingHandler) UpdatePairing(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req AdminUpdatePairingRequest
	if err := decodeJSONBody(w, r, &req, 32<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	status, err := parseAgencyPairingStatus(req.Status)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
		return
	}

	pairing, err := h.pairingService.UpdatePairingStatus(adminID, chi.URLParam(r, "id"), status, req.Notes, clientIP(r))
	if err != nil {
		h.writeAdminPairingError(w, err)
		return
	}

	items, err := h.mapAdminPairings([]*domain.AgencyPairing{pairing})
	if err != nil || len(items) == 0 {
		_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "pairing updated"})
		return
	}
	_ = utils.WriteJSON(w, http.StatusOK, map[string]AdminPairingResponse{"pairing": items[0]})
}

func (h *AdminPairingHandler) GetAgencyPairings(w http.ResponseWriter, r *http.Request) {
	pairings, err := h.pairingService.ListAgencyPairings(chi.URLParam(r, "id"))
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	items, err := h.mapAdminPairings(pairings)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	_ = utils.WriteJSON(w, http.StatusOK, map[string][]AdminPairingResponse{"pairings": items})
}

func (h *AdminPairingHandler) mapAdminPairings(pairings []*domain.AgencyPairing) ([]AdminPairingResponse, error) {
	items := make([]AdminPairingResponse, 0, len(pairings))
	for _, pairing := range pairings {
		if pairing == nil {
			continue
		}
		ethiopianUser, err := h.userRepository.GetByID(pairing.EthiopianUserID)
		if err != nil {
			return nil, err
		}
		foreignUser, err := h.userRepository.GetByID(pairing.ForeignUserID)
		if err != nil {
			return nil, err
		}

		item := AdminPairingResponse{
			ID:              pairing.ID,
			Status:          string(pairing.Status),
			EthiopianAgency: mapPairingAgencySummary(ethiopianUser),
			ForeignAgency:   mapPairingAgencySummary(foreignUser),
		}
		if pairing.ApprovedAt != nil {
			item.ApprovedAt = pairing.ApprovedAt.UTC().Format(time.RFC3339)
		}
		if pairing.Notes != nil {
			item.Notes = *pairing.Notes
		}
		items = append(items, item)
	}
	return items, nil
}

func parseAgencyPairingStatus(value string) (domain.AgencyPairingStatus, error) {
	switch strings.TrimSpace(value) {
	case string(domain.AgencyPairingActive):
		return domain.AgencyPairingActive, nil
	case string(domain.AgencyPairingSuspended):
		return domain.AgencyPairingSuspended, nil
	case string(domain.AgencyPairingEnded):
		return domain.AgencyPairingEnded, nil
	default:
		return "", errors.New("invalid pairing status")
	}
}

func (h *AdminPairingHandler) writeAdminPairingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrPairingAlreadyExists):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "pairing already exists"})
	case errors.Is(err, service.ErrInvalidPairingParticipants):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid pairing participants"})
	case errors.Is(err, service.ErrPairingNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "pairing not found"})
	default:
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
