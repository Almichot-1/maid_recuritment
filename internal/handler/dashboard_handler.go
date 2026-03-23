package handler

import (
	"errors"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type DashboardStatsResponse struct {
	TotalCandidates     int64 `json:"totalCandidates"`
	AvailableCandidates int64 `json:"availableCandidates"`
	SelectedCandidates  int64 `json:"selectedCandidates"`
	InProgress          int64 `json:"inProgress"`
	Approved            int64 `json:"approved"`
	ActiveSelections    int64 `json:"activeSelections"`
}

type DashboardHandler struct {
	candidateRepository    *repository.GormCandidateRepository
	selectionRepository    *repository.GormSelectionRepository
	notificationRepository *repository.GormNotificationRepository
	pairingService         *service.PairingService
}

func NewDashboardHandler(candidateRepository *repository.GormCandidateRepository, selectionRepository *repository.GormSelectionRepository, notificationRepository *repository.GormNotificationRepository, pairingService *service.PairingService) *DashboardHandler {
	return &DashboardHandler{
		candidateRepository:    candidateRepository,
		selectionRepository:    selectionRepository,
		notificationRepository: notificationRepository,
		pairingService:         pairingService,
	}
}

func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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
	if h.pairingService == nil {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "workspace access is not configured"})
		return
	}
	pairingID, _ := middleware.PairingIDFromContext(r.Context())
	pairing, err := h.pairingService.ResolveActivePairing(userID, role, pairingID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPairingRequired):
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "select a partner workspace to continue"})
		case errors.Is(err, service.ErrNoActivePairings), errors.Is(err, service.ErrPairingAccessDenied):
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		default:
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	stats := DashboardStatsResponse{}
	candidateDB := h.candidateRepository.DB()
	selectionDB := h.selectionRepository.DB()

	switch strings.TrimSpace(role) {
	case string(domain.EthiopianAgent):
		sharedCandidateBase := candidateDB.Model(&domain.Candidate{}).
			Joins("JOIN candidate_pair_shares ON candidate_pair_shares.candidate_id = candidates.id AND candidate_pair_shares.pairing_id = ? AND candidate_pair_shares.is_active = ?", pairing.ID, true).
			Where("candidates.created_by = ?", userID).
			Distinct("candidates.id")
		sharedCandidateBase.Count(&stats.TotalCandidates)
		sharedCandidateBase.Session(&gorm.Session{}).Where("candidates.status = ?", domain.CandidateStatusAvailable).Count(&stats.AvailableCandidates)
		sharedCandidateBase.Session(&gorm.Session{}).Where("candidates.status = ?", domain.CandidateStatusLocked).Count(&stats.SelectedCandidates)
		sharedCandidateBase.Session(&gorm.Session{}).Where("candidates.status IN ?", []domain.CandidateStatus{domain.CandidateStatusUnderReview, domain.CandidateStatusApproved, domain.CandidateStatusInProgress}).Count(&stats.InProgress)
		sharedCandidateBase.Session(&gorm.Session{}).Where("candidates.status = ?", domain.CandidateStatusApproved).Count(&stats.Approved)
		selectionDB.Model(&domain.Selection{}).
			Where("pairing_id = ? AND status = ?", pairing.ID, domain.SelectionPending).
			Count(&stats.ActiveSelections)
	case string(domain.ForeignAgent):
		sharedCandidateBase := candidateDB.Model(&domain.Candidate{}).
			Joins("JOIN candidate_pair_shares ON candidate_pair_shares.candidate_id = candidates.id AND candidate_pair_shares.pairing_id = ? AND candidate_pair_shares.is_active = ?", pairing.ID, true).
			Where("candidates.status = ?", domain.CandidateStatusAvailable).
			Distinct("candidates.id")
		sharedCandidateBase.Count(&stats.AvailableCandidates)
		selectionDB.Model(&domain.Selection{}).Where("selected_by = ? AND pairing_id = ? AND status = ?", userID, pairing.ID, domain.SelectionPending).Count(&stats.ActiveSelections)
		candidateDB.Model(&domain.Candidate{}).
			Joins("JOIN selections ON selections.candidate_id = candidates.id").
			Where("selections.selected_by = ? AND selections.pairing_id = ? AND candidates.status = ?", userID, pairing.ID, domain.CandidateStatusApproved).
			Count(&stats.Approved)
		candidateDB.Model(&domain.Candidate{}).
			Joins("JOIN selections ON selections.candidate_id = candidates.id").
			Where("selections.selected_by = ? AND selections.pairing_id = ? AND candidates.status = ?", userID, pairing.ID, domain.CandidateStatusInProgress).
			Count(&stats.InProgress)
		stats.TotalCandidates = stats.AvailableCandidates
	default:
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, stats)
}
