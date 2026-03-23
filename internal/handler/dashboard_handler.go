package handler

import (
	"errors"
	"net/http"
	"strings"

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

type dashboardStatsRow struct {
	TotalCandidates     int64 `gorm:"column:total_candidates"`
	AvailableCandidates int64 `gorm:"column:available_candidates"`
	SelectedCandidates  int64 `gorm:"column:selected_candidates"`
	InProgress          int64 `gorm:"column:in_progress"`
	Approved            int64 `gorm:"column:approved"`
	ActiveSelections    int64 `gorm:"column:active_selections"`
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
		row := dashboardStatsRow{}
		if err := candidateDB.Raw(`
			SELECT
				COUNT(DISTINCT c.id) AS total_candidates,
				COUNT(DISTINCT CASE WHEN c.status = ? THEN c.id END) AS available_candidates,
				COUNT(DISTINCT CASE WHEN c.status = ? THEN c.id END) AS selected_candidates,
				COUNT(DISTINCT CASE WHEN c.status IN (?, ?, ?) THEN c.id END) AS in_progress,
				COUNT(DISTINCT CASE WHEN c.status = ? THEN c.id END) AS approved
			FROM candidates c
			JOIN candidate_pair_shares cps
				ON cps.candidate_id = c.id
				AND cps.pairing_id = ?
				AND cps.is_active = TRUE
			WHERE c.created_by = ?
				AND c.deleted_at IS NULL
		`,
			domain.CandidateStatusAvailable,
			domain.CandidateStatusLocked,
			domain.CandidateStatusUnderReview,
			domain.CandidateStatusApproved,
			domain.CandidateStatusInProgress,
			domain.CandidateStatusApproved,
			pairing.ID,
			userID,
		).Scan(&row).Error; err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		stats.TotalCandidates = row.TotalCandidates
		stats.AvailableCandidates = row.AvailableCandidates
		stats.SelectedCandidates = row.SelectedCandidates
		stats.InProgress = row.InProgress
		stats.Approved = row.Approved
		if err := selectionDB.Model(&domain.Selection{}).
			Where("pairing_id = ? AND status = ?", pairing.ID, domain.SelectionPending).
			Count(&stats.ActiveSelections).Error; err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
	case string(domain.ForeignAgent):
		row := dashboardStatsRow{}
		if err := candidateDB.Raw(`
			SELECT
				COUNT(DISTINCT CASE WHEN c.status = ? THEN c.id END) AS available_candidates,
				COUNT(DISTINCT CASE WHEN c.status = ? AND s.id IS NOT NULL THEN c.id END) AS approved,
				COUNT(DISTINCT CASE WHEN c.status = ? AND s.id IS NOT NULL THEN c.id END) AS in_progress
			FROM candidates c
			JOIN candidate_pair_shares cps
				ON cps.candidate_id = c.id
				AND cps.pairing_id = ?
				AND cps.is_active = TRUE
			LEFT JOIN selections s
				ON s.candidate_id = c.id
				AND s.selected_by = ?
				AND s.pairing_id = ?
			WHERE c.deleted_at IS NULL
		`,
			domain.CandidateStatusAvailable,
			domain.CandidateStatusApproved,
			domain.CandidateStatusInProgress,
			pairing.ID,
			userID,
			pairing.ID,
		).Scan(&row).Error; err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		stats.AvailableCandidates = row.AvailableCandidates
		stats.Approved = row.Approved
		stats.InProgress = row.InProgress
		if err := selectionDB.Model(&domain.Selection{}).
			Where("selected_by = ? AND pairing_id = ? AND status = ?", userID, pairing.ID, domain.SelectionPending).
			Count(&stats.ActiveSelections).Error; err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		stats.TotalCandidates = stats.AvailableCandidates
	default:
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, stats)
}
