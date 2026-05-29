package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

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

// ── Smart-alerts response types ───────────────────────────────────────────────

// ExpiringSelectionAlert describes a pending selection that is about to expire.
type ExpiringSelectionAlert struct {
	SelectionID   string `json:"selection_id"`
	CandidateID   string `json:"candidate_id"`
	CandidateName string `json:"candidate_name"`
	ExpiresAt     string `json:"expires_at"`
	// HoursRemaining is negative when already expired.
	HoursRemaining float64 `json:"hours_remaining"`
	UrgencyLevel   string  `json:"urgency_level"` // "critical" | "high" | "medium"
}

// PassportExpiryAlert describes a candidate whose passport is expiring soon.
type PassportExpiryAlert struct {
	CandidateID    string `json:"candidate_id"`
	CandidateName  string `json:"candidate_name"`
	PassportNumber string `json:"passport_number"`
	ExpiryDate     string `json:"expiry_date"`
	DaysRemaining  int    `json:"days_remaining"`
	UrgencyLevel   string `json:"urgency_level"` // "expired" | "critical" | "warning" | "notice"
}

// FlightUpdateItem describes a candidate who is in a ticket / arrival step.
type FlightUpdateItem struct {
	CandidateID   string `json:"candidate_id"`
	CandidateName string `json:"candidate_name"`
	CurrentStep   string `json:"current_step"`
	StepStatus    string `json:"step_status"`
	UpdatedAt     string `json:"updated_at"`
}

// SmartAlertsResponse bundles all three alert categories.
type SmartAlertsResponse struct {
	ExpiringSelections []ExpiringSelectionAlert `json:"expiring_selections"`
	PassportExpiry     []PassportExpiryAlert    `json:"passport_expiry"`
	FlightUpdates      []FlightUpdateItem       `json:"flight_updates"`
}

type DashboardHandler struct {
	candidateRepository    *repository.GormCandidateRepository
	selectionRepository    *repository.GormSelectionRepository
	notificationRepository *repository.GormNotificationRepository
	passportRepository     domain.PassportDataRepository
	statusStepRepository   domain.StatusStepRepository
	pairingService         *service.PairingService
}

func NewDashboardHandler(
	candidateRepository *repository.GormCandidateRepository,
	selectionRepository *repository.GormSelectionRepository,
	notificationRepository *repository.GormNotificationRepository,
	pairingService *service.PairingService,
) *DashboardHandler {
	return &DashboardHandler{
		candidateRepository:    candidateRepository,
		selectionRepository:    selectionRepository,
		notificationRepository: notificationRepository,
		pairingService:         pairingService,
	}
}

// SetPassportRepository injects the passport data repository so smart alerts
// can include passport-expiry information.
func (h *DashboardHandler) SetPassportRepository(repo domain.PassportDataRepository) {
	h.passportRepository = repo
}

// SetStatusStepRepository injects the status step repository so smart alerts
// can surface candidates in flight-related tracking steps.
func (h *DashboardHandler) SetStatusStepRepository(repo domain.StatusStepRepository) {
	h.statusStepRepository = repo
}

// GetSmartAlerts handles GET /dashboard/smart-alerts
// Returns three categories:
//  1. Pending selections expiring within the next 24 hours
//  2. Candidates whose passports expire within 6 months
//  3. Candidates currently in a flight-related tracking step
func (h *DashboardHandler) GetSmartAlerts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	role, ok := middleware.RoleFromContext(r.Context())
	if !ok || strings.TrimSpace(role) != string(domain.EthiopianAgent) {
		// Smart alerts are only meaningful for Ethiopian agents.
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	resp := SmartAlertsResponse{
		ExpiringSelections: make([]ExpiringSelectionAlert, 0),
		PassportExpiry:     make([]PassportExpiryAlert, 0),
		FlightUpdates:      make([]FlightUpdateItem, 0),
	}

	candidateDB := h.candidateRepository.DB()
	selectionDB := h.selectionRepository.DB()

	// ── 1. Expiring selections (next 24 h) ───────────────────────────────────
	cutoff := time.Now().UTC().Add(24 * time.Hour)
	rawSelections := make([]*domain.Selection, 0)
	selectionDB.
		Where("status = ? AND expires_at <= ? AND expires_at > NOW()", domain.SelectionPending, cutoff).
		Order("expires_at ASC").
		Find(&rawSelections)

	for _, sel := range rawSelections {
		if sel == nil {
			continue
		}
		// Only show selections for candidates owned by this agent.
		var cand domain.Candidate
		if err := candidateDB.Where("id = ? AND created_by = ?", sel.CandidateID, userID).First(&cand).Error; err != nil {
			continue
		}
		hoursLeft := time.Until(sel.ExpiresAt).Hours()
		urgency := "medium"
		if hoursLeft < 1 {
			urgency = "critical"
		} else if hoursLeft < 6 {
			urgency = "high"
		}
		resp.ExpiringSelections = append(resp.ExpiringSelections, ExpiringSelectionAlert{
			SelectionID:    sel.ID,
			CandidateID:    sel.CandidateID,
			CandidateName:  cand.FullName,
			ExpiresAt:      sel.ExpiresAt.UTC().Format(time.RFC3339),
			HoursRemaining: hoursLeft,
			UrgencyLevel:   urgency,
		})
	}

	// ── 2. Passport expiry alerts (next 6 months) ────────────────────────────
	if h.passportRepository != nil {
		// 6 months ≈ 180 days; fetch any not yet warned at the 6-month level.
		// For the dashboard we just return ALL expiring within 180 days
		// regardless of warning flags – the flags only control email dispatch.
		passports := make([]*domain.PassportData, 0)
		cutoffPassport := time.Now().UTC().AddDate(0, 6, 0)
		// Use the repository method with flagBit=0 so all rows within the window are returned.
		var passportErr error
		passports, passportErr = h.passportRepository.GetExpiringPassports(180, 0)
		if passportErr == nil {
			for _, pd := range passports {
				if pd == nil || pd.ExpiryDate == nil {
					continue
				}
				// Scope to candidates owned by this agent.
				var cand domain.Candidate
				if err := candidateDB.Where("id = ? AND created_by = ?", pd.CandidateID, userID).First(&cand).Error; err != nil {
					continue
				}
				// Re-check the date is truly within 6 months (repository returns ≤ cutoff).
				if pd.ExpiryDate.After(cutoffPassport) {
					continue
				}
				days := pd.DaysUntilExpiry()
				urgency := "notice"
				if days <= 0 {
					urgency = "expired"
				} else if days <= 30 {
					urgency = "critical"
				} else if days <= 90 {
					urgency = "warning"
				}
				resp.PassportExpiry = append(resp.PassportExpiry, PassportExpiryAlert{
					CandidateID:    pd.CandidateID,
					CandidateName:  cand.FullName,
					PassportNumber: pd.PassportNumber,
					ExpiryDate:     pd.ExpiryDate.Format("2006-01-02"),
					DaysRemaining:  days,
					UrgencyLevel:   urgency,
				})
			}
		}
	}

	// ── 3. Flight updates ────────────────────────────────────────────────────
	// Return candidates owned by this agent who are in a ticket / arrival step.
	if h.statusStepRepository != nil {
		flightSteps := []string{
			domain.TicketPending,
			domain.TicketBooked,
			domain.TicketConfirmed,
			domain.Arrived,
		}

		// Fetch all in-progress candidates owned by this agent.
		inProgressCandidates := make([]*domain.Candidate, 0)
		candidateDB.
			Where("created_by = ? AND status IN ?", userID,
				[]domain.CandidateStatus{domain.CandidateStatusInProgress, domain.CandidateStatusCompleted}).
			Find(&inProgressCandidates)

		for _, cand := range inProgressCandidates {
			if cand == nil {
				continue
			}
			steps, err := h.statusStepRepository.GetByCandidateID(cand.ID)
			if err != nil {
				continue
			}
			for _, step := range steps {
				if step == nil {
					continue
				}
				isFlightStep := false
				for _, fs := range flightSteps {
					if strings.EqualFold(strings.TrimSpace(step.StepName), fs) {
						isFlightStep = true
						break
					}
				}
				if !isFlightStep {
					continue
				}
				// Only surface steps that are in_progress or completed.
				if step.StepStatus == domain.Pending {
					continue
				}
				resp.FlightUpdates = append(resp.FlightUpdates, FlightUpdateItem{
					CandidateID:   cand.ID,
					CandidateName: cand.FullName,
					CurrentStep:   step.StepName,
					StepStatus:    string(step.StepStatus),
					UpdatedAt:     step.UpdatedAt.UTC().Format(time.RFC3339),
				})
			}
		}
	}

	_ = utils.WriteJSON(w, http.StatusOK, resp)
}

type DashboardHomeResponse struct {
	Stats              DashboardStatsResponse       `json:"stats"`
	PendingActions     DashboardPendingActions      `json:"pending_actions"`
	RecentCandidates   []domain.Candidate           `json:"recent_candidates,omitempty"`
	AvailableCandidates []domain.Candidate          `json:"available_candidates,omitempty"`
	ActiveSelections   []DashboardSelectionPreview  `json:"active_selections,omitempty"`
	ApprovedSelections []DashboardSelectionPreview  `json:"approved_selections,omitempty"`
}

type DashboardPendingActions struct {
	IncompleteProfiles int64 `json:"incomplete_profiles"`
	ActiveSelections   int64 `json:"active_selections"`
}

type DashboardSelectionPreview struct {
	ID            string  `json:"id"`
	CandidateID   string  `json:"candidate_id"`
	CandidateName string  `json:"candidate_name"`
	Status        string  `json:"status"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

func (h *DashboardHandler) GetHome(w http.ResponseWriter, r *http.Request) {
	recorder := httptest.NewRecorder()
	h.GetStats(recorder, r)
	if recorder.Code != http.StatusOK {
		w.WriteHeader(recorder.Code)
		_, _ = w.Write(recorder.Body.Bytes())
		return
	}

	var stats DashboardStatsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &stats); err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, DashboardHomeResponse{
		Stats: stats,
		PendingActions: DashboardPendingActions{
			IncompleteProfiles: 0,
			ActiveSelections:   stats.ActiveSelections,
		},
		RecentCandidates:    []domain.Candidate{},
		AvailableCandidates: []domain.Candidate{},
		ActiveSelections:    []DashboardSelectionPreview{},
		ApprovedSelections:  []DashboardSelectionPreview{},
	})
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
