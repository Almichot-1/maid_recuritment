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

type AdminDashboardHandler struct {
	userRepository      *repository.GormUserRepository
	candidateRepository *repository.GormCandidateRepository
	selectionRepository *repository.GormSelectionRepository
}

func NewAdminDashboardHandler(userRepository *repository.GormUserRepository, candidateRepository *repository.GormCandidateRepository, selectionRepository *repository.GormSelectionRepository) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		userRepository:      userRepository,
		candidateRepository: candidateRepository,
		selectionRepository: selectionRepository,
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
