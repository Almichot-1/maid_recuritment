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

type DashboardPendingActionsResponse struct {
	IncompleteProfiles int64 `json:"incompleteProfiles"`
	ActiveSelections   int64 `json:"activeSelections"`
}

type DashboardSelectionPreviewResponse struct {
	ID            string `json:"id"`
	CandidateID   string `json:"candidate_id"`
	CandidateName string `json:"candidate_name"`
	Status        string `json:"status"`
	ExpiresAt     string `json:"expires_at,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type DashboardHomeResponse struct {
	Stats               DashboardStatsResponse              `json:"stats"`
	PendingActions      DashboardPendingActionsResponse     `json:"pending_actions"`
	RecentCandidates    []CandidateResponse                 `json:"recent_candidates,omitempty"`
	AvailableCandidates []CandidateResponse                 `json:"available_candidates,omitempty"`
	ActiveSelections    []DashboardSelectionPreviewResponse `json:"active_selections,omitempty"`
	ApprovedSelections  []DashboardSelectionPreviewResponse `json:"approved_selections,omitempty"`
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
	userID, role, pairing, ok := h.resolveDashboardContext(w, r)
	if !ok {
		return
	}

	stats, err := h.loadStats(userID, role, pairing)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, stats)
}

func (h *DashboardHandler) GetHome(w http.ResponseWriter, r *http.Request) {
	userID, role, pairing, ok := h.resolveDashboardContext(w, r)
	if !ok {
		return
	}

	stats, err := h.loadStats(userID, role, pairing)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	home := DashboardHomeResponse{
		Stats: stats,
		PendingActions: DashboardPendingActionsResponse{
			ActiveSelections: stats.ActiveSelections,
		},
	}

	switch strings.TrimSpace(role) {
	case string(domain.EthiopianAgent):
		home.RecentCandidates, err = h.loadRecentCandidates(userID)
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		home.PendingActions.IncompleteProfiles, err = h.countIncompleteProfiles(userID)
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
	case string(domain.ForeignAgent):
		home.AvailableCandidates, err = h.loadAvailableCandidatesForWorkspace(userID, pairing.ID)
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		home.ActiveSelections, err = h.loadSelectionPreviews(userID, pairing.ID, domain.SelectionPending, 3)
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		home.ApprovedSelections, err = h.loadSelectionPreviews(userID, pairing.ID, domain.SelectionApproved, 2)
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
	default:
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, home)
}

func (h *DashboardHandler) resolveDashboardContext(w http.ResponseWriter, r *http.Request) (string, string, *domain.AgencyPairing, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return "", "", nil, false
	}
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok || strings.TrimSpace(role) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return "", "", nil, false
	}
	if h.pairingService == nil {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "workspace access is not configured"})
		return "", "", nil, false
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
		return "", "", nil, false
	}

	return userID, role, pairing, true
}

func (h *DashboardHandler) loadStats(userID, role string, pairing *domain.AgencyPairing) (DashboardStatsResponse, error) {
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
			return DashboardStatsResponse{}, err
		}
		stats.TotalCandidates = row.TotalCandidates
		stats.AvailableCandidates = row.AvailableCandidates
		stats.SelectedCandidates = row.SelectedCandidates
		stats.InProgress = row.InProgress
		stats.Approved = row.Approved
		if err := selectionDB.Model(&domain.Selection{}).
			Where("pairing_id = ? AND status = ?", pairing.ID, domain.SelectionPending).
			Count(&stats.ActiveSelections).Error; err != nil {
			return DashboardStatsResponse{}, err
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
			return DashboardStatsResponse{}, err
		}
		stats.AvailableCandidates = row.AvailableCandidates
		stats.Approved = row.Approved
		stats.InProgress = row.InProgress
		if err := selectionDB.Model(&domain.Selection{}).
			Where("selected_by = ? AND pairing_id = ? AND status = ?", userID, pairing.ID, domain.SelectionPending).
			Count(&stats.ActiveSelections).Error; err != nil {
			return DashboardStatsResponse{}, err
		}
		stats.TotalCandidates = stats.AvailableCandidates
	default:
		return DashboardStatsResponse{}, service.ErrForbidden
	}

	return stats, nil
}

func (h *DashboardHandler) loadRecentCandidates(userID string) ([]CandidateResponse, error) {
	candidates, err := h.candidateRepository.List(domain.CandidateFilters{
		CreatedBy: userID,
		Page:      1,
		PageSize:  5,
	})
	if err != nil {
		return nil, err
	}

	responses := make([]CandidateResponse, 0, len(candidates))
	for _, candidate := range candidates {
		responses = append(responses, mapCandidateResponse(candidate, nil))
	}

	return responses, nil
}

func (h *DashboardHandler) loadAvailableCandidatesForWorkspace(userID, pairingID string) ([]CandidateResponse, error) {
	candidates, err := h.candidateRepository.List(domain.CandidateFilters{
		PairingID:  pairingID,
		SharedOnly: true,
		Statuses:   []domain.CandidateStatus{domain.CandidateStatusAvailable},
		Page:       1,
		PageSize:   4,
	})
	if err != nil {
		return nil, err
	}

	responses := make([]CandidateResponse, 0, len(candidates))
	for _, candidate := range candidates {
		sanitizeCandidateForViewer(candidate, userID, string(domain.ForeignAgent))
		responses = append(responses, mapCandidateResponse(candidate, nil))
	}

	return responses, nil
}

func (h *DashboardHandler) countIncompleteProfiles(userID string) (int64, error) {
	var count int64
	err := h.candidateRepository.DB().Raw(`
		SELECT COUNT(*)
		FROM candidates c
		WHERE c.created_by = ?
			AND c.deleted_at IS NULL
			AND (
				NOT EXISTS (
					SELECT 1
					FROM documents d
					WHERE d.candidate_id = c.id
						AND d.document_type = ?
				)
				OR NOT EXISTS (
					SELECT 1
					FROM documents d
					WHERE d.candidate_id = c.id
						AND d.document_type = ?
				)
			)
	`, userID, domain.Passport, domain.Photo).Scan(&count).Error
	return count, err
}

func (h *DashboardHandler) loadSelectionPreviews(userID, pairingID string, status domain.SelectionStatus, limit int) ([]DashboardSelectionPreviewResponse, error) {
	selections, err := h.selectionRepository.GetBySelectedByAndPairing(userID, pairingID)
	if err != nil {
		return nil, err
	}

	previews := make([]DashboardSelectionPreviewResponse, 0, limit)
	for _, selection := range selections {
		if selection == nil || selection.Status != status {
			continue
		}

		candidate, err := h.candidateRepository.GetByID(selection.CandidateID)
		if err != nil {
			if errors.Is(err, repository.ErrCandidateNotFound) {
				continue
			}
			return nil, err
		}

		preview := DashboardSelectionPreviewResponse{
			ID:            selection.ID,
			CandidateID:   selection.CandidateID,
			CandidateName: candidate.FullName,
			Status:        string(selection.Status),
			CreatedAt:     selection.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		}
		if !selection.ExpiresAt.IsZero() {
			preview.ExpiresAt = selection.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		}

		previews = append(previews, preview)
		if len(previews) >= limit {
			break
		}
	}

	return previews, nil
}
