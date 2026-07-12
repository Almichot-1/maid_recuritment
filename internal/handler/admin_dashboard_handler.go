package handler

import (
	"net/http"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/pkg/utils"
)

type AdminDashboardStatsResponse struct {
	TotalAgencies         int64   `json:"total_agencies"`
	EthiopianAgencies     int64   `json:"ethiopian_agencies"`
	ForeignAgencies       int64   `json:"foreign_agencies"`
	PendingApprovals      int64   `json:"pending_approvals"`
	TotalCandidates       int64   `json:"total_candidates"`
	ActiveSelections      int64   `json:"active_selections"`
	CompletedRecruitments int64   `json:"completed_recruitments"`
	SuccessRate           float64 `json:"success_rate"`
}

type AdminProgressMetricsResponse struct {
	TotalInProgress int64 `json:"total_in_progress"`
	TotalCompleted  int64 `json:"total_completed"`
	TotalFailed     int64 `json:"total_failed"`
	StuckAtCOC      int64 `json:"stuck_at_coc"`
	StuckAtMedical  int64 `json:"stuck_at_medical"`
	StuckAtVisa     int64 `json:"stuck_at_visa"`
	StuckAtTicket   int64 `json:"stuck_at_ticket"`
}

type AdminDashboardHandler struct {
	userRepository            *repository.GormUserRepository
	candidateRepository       *repository.GormCandidateRepository
	selectionRepository       *repository.GormSelectionRepository
	selectionProgressRepo     *repository.GormSelectionProgressRepository
}

func NewAdminDashboardHandler(userRepository *repository.GormUserRepository, candidateRepository *repository.GormCandidateRepository, selectionRepository *repository.GormSelectionRepository, selectionProgressRepo *repository.GormSelectionProgressRepository) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		userRepository:        userRepository,
		candidateRepository:   candidateRepository,
		selectionRepository:   selectionRepository,
		selectionProgressRepo: selectionProgressRepo,
	}
}

func (h *AdminDashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := AdminDashboardStatsResponse{}

	userDB := h.userRepository.DB()
	candidateDB := h.candidateRepository.DB()
	selectionDB := h.selectionRepository.DB()

	userDB.Model(&domain.User{}).Where("role IN ?", []domain.UserRole{domain.EthiopianAgent, domain.ForeignAgent}).Count(&stats.TotalAgencies)
	userDB.Model(&domain.User{}).Where("role = ?", domain.EthiopianAgent).Count(&stats.EthiopianAgencies)
	userDB.Model(&domain.User{}).Where("role = ?", domain.ForeignAgent).Count(&stats.ForeignAgencies)
	userDB.Model(&domain.User{}).Where("account_status = ?", domain.AccountStatusPendingApproval).Count(&stats.PendingApprovals)

	candidateDB.Model(&domain.Candidate{}).Count(&stats.TotalCandidates)
	selectionDB.Model(&domain.Selection{}).Where("status = ?", domain.SelectionPending).Count(&stats.ActiveSelections)
	candidateDB.Model(&domain.Candidate{}).Where("status = ?", domain.CandidateStatusCompleted).Count(&stats.CompletedRecruitments)

	var totalSelections int64
	var approvedSelections int64
	selectionDB.Model(&domain.Selection{}).Count(&totalSelections)
	selectionDB.Model(&domain.Selection{}).Where("status = ?", domain.SelectionApproved).Count(&approvedSelections)
	if totalSelections > 0 {
		stats.SuccessRate = float64(approvedSelections) / float64(totalSelections)
	}

	_ = utils.WriteJSON(w, http.StatusOK, stats)
}

func (h *AdminDashboardHandler) GetProgressMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := AdminProgressMetricsResponse{}
	db := h.selectionProgressRepo.DB()

	db.Model(&domain.SelectionProgress{}).Count(&metrics.TotalInProgress)

	var totalCompleted int64
	db.Model(&domain.SelectionProgress{}).Where("coc_status = 'done' AND medical_status = 'done' AND visa_status = 'approved' AND ticket_status = 'confirmed' AND arrival_status = 'arrived'").Count(&totalCompleted)
	metrics.TotalCompleted = totalCompleted

	var totalFailed int64
	db.Model(&domain.SelectionProgress{}).Where("coc_status = 'failed' OR medical_status = 'failed' OR visa_status = 'rejected'").Count(&totalFailed)
	metrics.TotalFailed = totalFailed

	db.Model(&domain.SelectionProgress{}).Where("coc_status IN ('pending', 'in_progress') AND medical_status = 'pending' AND visa_status = 'pending' AND ticket_status = 'pending'").Count(&metrics.StuckAtCOC)
	db.Model(&domain.SelectionProgress{}).Where("coc_status = 'done' AND medical_status IN ('pending', 'in_progress') AND visa_status = 'pending'").Count(&metrics.StuckAtMedical)
	db.Model(&domain.SelectionProgress{}).Where("medical_status = 'done' AND visa_status IN ('pending', 'in_progress') AND ticket_status = 'pending'").Count(&metrics.StuckAtVisa)
	db.Model(&domain.SelectionProgress{}).Where("visa_status IN ('approved', 'in_progress') AND ticket_status IN ('pending', 'booked') AND arrival_status = 'not_arrived'").Count(&metrics.StuckAtTicket)

	_ = utils.WriteJSON(w, http.StatusOK, metrics)
}
