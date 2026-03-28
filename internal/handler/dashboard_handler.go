package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

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

type SmartAlertSelectionResponse struct {
	SelectionID    string `json:"selection_id"`
	CandidateID    string `json:"candidate_id"`
	CandidateName  string `json:"candidate_name"`
	ExpiresAt      string `json:"expires_at"`
	WarningLevel   string `json:"warning_level"`
	RemainingLabel string `json:"remaining_label"`
}

type SmartAlertPassportResponse struct {
	CandidateID    string `json:"candidate_id"`
	CandidateName  string `json:"candidate_name"`
	PassportNumber string `json:"passport_number"`
	ExpiryDate     string `json:"expiry_date"`
	WarningLevel   string `json:"warning_level"`
}

type SmartAlertMedicalResponse struct {
	CandidateID   string `json:"candidate_id"`
	CandidateName string `json:"candidate_name"`
	ExpiryDate    string `json:"expiry_date"`
	WarningLevel  string `json:"warning_level"`
}

type SmartAlertFlightResponse struct {
	CandidateID   string `json:"candidate_id"`
	CandidateName string `json:"candidate_name"`
	Stage         string `json:"stage"`
	UpdatedAt     string `json:"updated_at"`
	Status        string `json:"status"`
}

type DashboardSmartAlertsResponse struct {
	ExpiringSelections []SmartAlertSelectionResponse `json:"expiring_selections"`
	ExpiringPassports  []SmartAlertPassportResponse  `json:"expiring_passports"`
	ExpiringMedicals   []SmartAlertMedicalResponse   `json:"expiring_medicals"`
	FlightUpdates      []SmartAlertFlightResponse    `json:"flight_updates"`
	RecentlyArrived    []SmartAlertFlightResponse    `json:"recently_arrived"`
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
	passportRepository     *repository.GormPassportDataRepository
	medicalRepository      *repository.GormMedicalDataRepository
	statusStepRepository   domain.StatusStepRepository
}

type dashboardStatsRow struct {
	TotalCandidates     int64 `gorm:"column:total_candidates"`
	AvailableCandidates int64 `gorm:"column:available_candidates"`
	SelectedCandidates  int64 `gorm:"column:selected_candidates"`
	InProgress          int64 `gorm:"column:in_progress"`
	Approved            int64 `gorm:"column:approved"`
	ActiveSelections    int64 `gorm:"column:active_selections"`
}

func NewDashboardHandler(candidateRepository *repository.GormCandidateRepository, selectionRepository *repository.GormSelectionRepository, notificationRepository *repository.GormNotificationRepository, pairingService *service.PairingService, passportRepository *repository.GormPassportDataRepository, medicalRepository *repository.GormMedicalDataRepository, statusStepRepository domain.StatusStepRepository) *DashboardHandler {
	return &DashboardHandler{
		candidateRepository:    candidateRepository,
		selectionRepository:    selectionRepository,
		notificationRepository: notificationRepository,
		pairingService:         pairingService,
		passportRepository:     passportRepository,
		medicalRepository:      medicalRepository,
		statusStepRepository:   statusStepRepository,
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

func (h *DashboardHandler) GetSmartAlerts(w http.ResponseWriter, r *http.Request) {
	userID, role, pairing, ok := h.resolveDashboardContext(w, r)
	if !ok {
		return
	}
	if strings.TrimSpace(role) != string(domain.EthiopianAgent) {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	response, err := h.loadSmartAlerts(userID, pairing.ID)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, response)
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

func (h *DashboardHandler) loadSmartAlerts(userID, pairingID string) (*DashboardSmartAlertsResponse, error) {
	response := &DashboardSmartAlertsResponse{
		ExpiringSelections: make([]SmartAlertSelectionResponse, 0),
		ExpiringPassports:  make([]SmartAlertPassportResponse, 0),
		ExpiringMedicals:   make([]SmartAlertMedicalResponse, 0),
		FlightUpdates:      make([]SmartAlertFlightResponse, 0),
		RecentlyArrived:    make([]SmartAlertFlightResponse, 0),
	}

	type expiringSelectionRow struct {
		SelectionID   string    `gorm:"column:selection_id"`
		CandidateID   string    `gorm:"column:candidate_id"`
		CandidateName string    `gorm:"column:candidate_name"`
		ExpiresAt     time.Time `gorm:"column:expires_at"`
	}

	var expiringSelections []expiringSelectionRow
	if err := h.selectionRepository.DB().Raw(`
		SELECT
			s.id AS selection_id,
			s.candidate_id,
			c.full_name AS candidate_name,
			s.expires_at
		FROM selections s
		JOIN candidates c ON c.id = s.candidate_id
		WHERE c.created_by = ?
			AND s.pairing_id = ?
			AND s.status = ?
			AND s.expires_at > NOW()
			AND s.expires_at <= NOW() + INTERVAL '24 hours'
		ORDER BY s.expires_at ASC
		LIMIT 6
	`, userID, pairingID, domain.SelectionPending).Scan(&expiringSelections).Error; err != nil {
		return nil, err
	}
	for _, selection := range expiringSelections {
		level, remainingLabel := selectionWarningLabel(selection.ExpiresAt)
		response.ExpiringSelections = append(response.ExpiringSelections, SmartAlertSelectionResponse{
			SelectionID:    selection.SelectionID,
			CandidateID:    selection.CandidateID,
			CandidateName:  selection.CandidateName,
			ExpiresAt:      selection.ExpiresAt.UTC().Format(time.RFC3339),
			WarningLevel:   level,
			RemainingLabel: remainingLabel,
		})
	}

	if h.passportRepository != nil {
		type expiringPassportRow struct {
			CandidateID    string    `gorm:"column:candidate_id"`
			CandidateName  string    `gorm:"column:candidate_name"`
			PassportNumber string    `gorm:"column:passport_number"`
			ExpiryDate     time.Time `gorm:"column:expiry_date"`
		}
		var expiringPassports []expiringPassportRow
		if err := h.passportRepository.DB().Raw(`
			SELECT
				pd.candidate_id,
				c.full_name AS candidate_name,
				pd.passport_number,
				pd.expiry_date
			FROM passport_data pd
			JOIN candidates c ON c.id = pd.candidate_id
			JOIN candidate_pair_shares cps
				ON cps.candidate_id = c.id
				AND cps.pairing_id = ?
				AND cps.is_active = TRUE
			WHERE c.created_by = ?
				AND pd.expiry_date >= NOW()
				AND pd.expiry_date <= NOW() + INTERVAL '6 months'
			ORDER BY pd.expiry_date ASC
			LIMIT 6
		`, pairingID, userID).Scan(&expiringPassports).Error; err != nil {
			return nil, err
		}
		for _, passport := range expiringPassports {
			response.ExpiringPassports = append(response.ExpiringPassports, SmartAlertPassportResponse{
				CandidateID:    passport.CandidateID,
				CandidateName:  passport.CandidateName,
				PassportNumber: passport.PassportNumber,
				ExpiryDate:     passport.ExpiryDate.UTC().Format(time.RFC3339),
				WarningLevel:   passportWarningLabel(passport.ExpiryDate),
			})
		}
	}

	if h.medicalRepository != nil {
		type expiringMedicalRow struct {
			CandidateID   string    `gorm:"column:candidate_id"`
			CandidateName string    `gorm:"column:candidate_name"`
			ExpiryDate    time.Time `gorm:"column:expiry_date"`
		}
		var expiringMedicals []expiringMedicalRow
		if err := h.medicalRepository.DB().Raw(`
			SELECT
				md.candidate_id,
				c.full_name AS candidate_name,
				md.expiry_date
			FROM medical_data md
			JOIN candidates c ON c.id = md.candidate_id
			JOIN candidate_pair_shares cps
				ON cps.candidate_id = c.id
				AND cps.pairing_id = ?
				AND cps.is_active = TRUE
			WHERE c.created_by = ?
				AND md.expiry_date >= NOW()
				AND md.expiry_date <= NOW() + INTERVAL '30 days'
			ORDER BY md.expiry_date ASC
			LIMIT 6
		`, pairingID, userID).Scan(&expiringMedicals).Error; err != nil {
			return nil, err
		}
		for _, medical := range expiringMedicals {
			response.ExpiringMedicals = append(response.ExpiringMedicals, SmartAlertMedicalResponse{
				CandidateID:   medical.CandidateID,
				CandidateName: medical.CandidateName,
				ExpiryDate:    medical.ExpiryDate.UTC().Format(time.RFC3339),
				WarningLevel:  medicalWarningLabel(medical.ExpiryDate),
			})
		}
	}

	candidates, err := h.candidateRepository.List(domain.CandidateFilters{
		CreatedBy:  userID,
		PairingID:  pairingID,
		SharedOnly: true,
		Statuses:   []domain.CandidateStatus{domain.CandidateStatusInProgress, domain.CandidateStatusCompleted},
		Page:       1,
		PageSize:   12,
	})
	if err != nil {
		return nil, err
	}

	for _, candidate := range candidates {
		if candidate == nil || h.statusStepRepository == nil {
			continue
		}
		steps, err := h.statusStepRepository.GetByCandidateID(candidate.ID)
		if err != nil {
			return nil, err
		}
		stage, updatedAt, arrived := deriveFlightStage(steps)
		if stage == "" {
			continue
		}
		record := SmartAlertFlightResponse{
			CandidateID:   candidate.ID,
			CandidateName: candidate.FullName,
			Stage:         stage,
			UpdatedAt:     updatedAt.UTC().Format(time.RFC3339),
			Status:        string(candidate.Status),
		}
		if arrived {
			response.RecentlyArrived = append(response.RecentlyArrived, record)
		} else {
			response.FlightUpdates = append(response.FlightUpdates, record)
		}
	}

	return response, nil
}

func selectionWarningLabel(expiresAt time.Time) (string, string) {
	remaining := time.Until(expiresAt.UTC())
	switch {
	case remaining <= time.Hour:
		return "critical", "Less than 1 hour"
	case remaining <= 6*time.Hour:
		return "high", "Less than 6 hours"
	default:
		return "medium", "Less than 24 hours"
	}
}

func passportWarningLabel(expiryDate time.Time) string {
	remaining := time.Until(expiryDate.UTC())
	switch {
	case remaining <= 30*24*time.Hour:
		return "critical"
	case remaining <= 90*24*time.Hour:
		return "high"
	default:
		return "medium"
	}
}

func medicalWarningLabel(expiryDate time.Time) string {
	remaining := time.Until(expiryDate.UTC())
	switch {
	case remaining <= 7*24*time.Hour:
		return "critical"
	case remaining <= 14*24*time.Hour:
		return "high"
	default:
		return "medium"
	}
}

func deriveFlightStage(steps []*domain.StatusStep) (string, time.Time, bool) {
	order := []string{domain.TicketPending, domain.TicketBooked, domain.TicketConfirmed, domain.Arrived}
	for index := len(order) - 1; index >= 0; index-- {
		name := order[index]
		for _, step := range steps {
			if step == nil || strings.TrimSpace(step.StepName) != name {
				continue
			}
			if step.StepStatus == domain.Completed || step.StepStatus == domain.InProgress {
				return name, step.UpdatedAt, name == domain.Arrived
			}
		}
	}
	return "", time.Time{}, false
}
