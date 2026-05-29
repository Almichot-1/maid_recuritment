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

type AdminAgencySummary struct {
	ID               string `json:"id"`
	CompanyName      string `json:"company_name"`
	ContactPerson    string `json:"contact_person"`
	Email            string `json:"email"`
	Role             string `json:"role"`
	AccountStatus    string `json:"account_status"`
	RegistrationDate string `json:"registration_date"`
	TotalCandidates  int64  `json:"total_candidates"`
	TotalSelections  int64  `json:"total_selections"`
}

type AdminAgencyActivitySummary struct {
	TotalCandidates       int64 `json:"total_candidates"`
	ActiveCandidates      int64 `json:"active_candidates"`
	CompletedRecruitments int64 `json:"completed_recruitments"`
	TotalSelections       int64 `json:"total_selections"`
	ApprovedSelections    int64 `json:"approved_selections"`
	ActiveRecruitments    int64 `json:"active_recruitments"`
}

type AdminRecentActivityItem struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	OccurredAt string `json:"occurred_at"`
}

type AdminAgencyDetailResponse struct {
	Agency             AdminAgencySummary         `json:"agency"`
	ApprovalStatus     string                     `json:"approval_status"`
	RejectionReason    string                     `json:"rejection_reason,omitempty"`
	AdminNotes         string                     `json:"admin_notes,omitempty"`
	SubmittedDocuments []map[string]string        `json:"submitted_documents"`
	ActivitySummary    AdminAgencyActivitySummary `json:"activity_summary"`
	RecentActivity     []AdminRecentActivityItem  `json:"recent_activity"`
}

type AdminRejectAgencyRequest struct {
	Reason string `json:"reason"`
	Notes  string `json:"notes"`
}

type AdminUpdateAgencyStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type AdminAgencyHandler struct {
	userRepository      *repository.GormUserRepository
	approvalRepository  domain.AgencyApprovalRequestRepository
	approvalService     *service.AgencyApprovalService
	candidateRepository *repository.GormCandidateRepository
	selectionRepository *repository.GormSelectionRepository
}

func NewAdminAgencyHandler(
	userRepository *repository.GormUserRepository,
	approvalRepository domain.AgencyApprovalRequestRepository,
	approvalService *service.AgencyApprovalService,
	candidateRepository *repository.GormCandidateRepository,
	selectionRepository *repository.GormSelectionRepository,
) *AdminAgencyHandler {
	return &AdminAgencyHandler{
		userRepository:      userRepository,
		approvalRepository:  approvalRepository,
		approvalService:     approvalService,
		candidateRepository: candidateRepository,
		selectionRepository: selectionRepository,
	}
}

func (h *AdminAgencyHandler) GetPendingAgencies(w http.ResponseWriter, r *http.Request) {
	roleFilter := strings.TrimSpace(r.URL.Query().Get("role"))
	requests, err := h.approvalRepository.ListByStatus(domain.AgencyApprovalPending)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	agencies := make([]AdminAgencySummary, 0, len(requests))
	for _, request := range requests {
		user, err := h.userRepository.GetByID(request.AgencyID)
		if err != nil {
			continue
		}
		if !matchesAgencyRoleFilter(roleFilter, user.Role) {
			continue
		}
		summary, err := h.buildAgencySummary(user)
		if err != nil {
			continue
		}
		agencies = append(agencies, summary)
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string][]AdminAgencySummary{"agencies": agencies})
}

func (h *AdminAgencyHandler) GetAgencies(w http.ResponseWriter, r *http.Request) {
	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	roleFilter := strings.TrimSpace(r.URL.Query().Get("role"))
	search := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("search")))

	users := make([]*domain.User, 0)
	query := h.userRepository.DB().Model(&domain.User{}).Where("role IN ?", []domain.UserRole{domain.EthiopianAgent, domain.ForeignAgent})
	if statusFilter != "" {
		query = query.Where("account_status = ?", statusFilter)
	}
	if roleFilter != "" {
		query = query.Where("role = ?", roleFilter)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("LOWER(email) LIKE ? OR LOWER(full_name) LIKE ? OR LOWER(company_name) LIKE ?", like, like, like)
	}
	if err := query.Order("created_at DESC").Find(&users).Error; err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	agencies := make([]AdminAgencySummary, 0, len(users))
	for _, user := range users {
		summary, err := h.buildAgencySummary(user)
		if err != nil {
			continue
		}
		agencies = append(agencies, summary)
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string][]AdminAgencySummary{"agencies": agencies})
}

func (h *AdminAgencyHandler) GetAgency(w http.ResponseWriter, r *http.Request) {
	agencyID := chi.URLParam(r, "id")
	user, err := h.userRepository.GetByID(agencyID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "agency not found"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	summary, err := h.buildAgencySummary(user)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	request, err := h.approvalRepository.GetByAgencyID(user.ID)
	if err != nil && !errors.Is(err, repository.ErrAgencyApprovalNotFound) {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	detail := AdminAgencyDetailResponse{
		Agency:             summary,
		SubmittedDocuments: []map[string]string{},
		ActivitySummary:    h.buildActivitySummary(user),
		RecentActivity:     h.buildRecentActivity(user),
	}
	if request != nil {
		detail.ApprovalStatus = string(request.Status)
		if request.RejectionReason != nil {
			detail.RejectionReason = *request.RejectionReason
		}
		if request.AdminNotes != nil {
			detail.AdminNotes = *request.AdminNotes
		}
	} else {
		detail.ApprovalStatus = string(domain.AgencyApprovalPending)
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]AdminAgencyDetailResponse{"agency": detail})
}

func (h *AdminAgencyHandler) ApproveAgency(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	agencyID := chi.URLParam(r, "id")
	if err := h.approvalService.ApproveAgency(adminID, agencyID, clientIP(r)); err != nil {
		h.writeAgencyServiceError(w, err)
		return
	}
	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "agency approved"})
}

func (h *AdminAgencyHandler) RejectAgency(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req AdminRejectAgencyRequest
	if err := decodeJSONBody(w, r, &req, 32<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(req.Reason) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "rejection reason is required"})
		return
	}
	agencyID := chi.URLParam(r, "id")
	if err := h.approvalService.RejectAgency(adminID, agencyID, req.Reason, req.Notes, clientIP(r)); err != nil {
		h.writeAgencyServiceError(w, err)
		return
	}
	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "agency rejected"})
}

func (h *AdminAgencyHandler) UpdateAgencyStatus(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AdminIDFromContext(r.Context())
	if !ok {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req AdminUpdateAgencyStatusRequest
	if err := decodeJSONBody(w, r, &req, 32<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	status, err := parseAgencyAccountStatus(req.Status)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
		return
	}
	if err := h.approvalService.UpdateAgencyStatus(adminID, chi.URLParam(r, "id"), status, req.Reason, clientIP(r)); err != nil {
		h.writeAgencyServiceError(w, err)
		return
	}
	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "agency status updated"})
}

func (h *AdminAgencyHandler) buildAgencySummary(user *domain.User) (AdminAgencySummary, error) {
	summary := AdminAgencySummary{
		ID:               user.ID,
		CompanyName:      user.CompanyName,
		ContactPerson:    user.FullName,
		Email:            user.Email,
		Role:             string(user.Role),
		AccountStatus:    string(user.AccountStatus),
		RegistrationDate: user.CreatedAt.UTC().Format(time.RFC3339),
	}

	switch user.Role {
	case domain.EthiopianAgent:
		h.candidateRepository.DB().Model(&domain.Candidate{}).Where("created_by = ?", user.ID).Count(&summary.TotalCandidates)
	case domain.ForeignAgent:
		h.selectionRepository.DB().Model(&domain.Selection{}).Where("selected_by = ?", user.ID).Count(&summary.TotalSelections)
	}

	return summary, nil
}

func (h *AdminAgencyHandler) buildActivitySummary(user *domain.User) AdminAgencyActivitySummary {
	summary := AdminAgencyActivitySummary{}
	switch user.Role {
	case domain.EthiopianAgent:
		db := h.candidateRepository.DB().Model(&domain.Candidate{})
		db.Where("created_by = ?", user.ID).Count(&summary.TotalCandidates)
		db.Where("created_by = ? AND status IN ?", user.ID, []domain.CandidateStatus{domain.CandidateStatusAvailable, domain.CandidateStatusLocked, domain.CandidateStatusUnderReview, domain.CandidateStatusApproved, domain.CandidateStatusInProgress}).Count(&summary.ActiveCandidates)
		db.Where("created_by = ? AND status = ?", user.ID, domain.CandidateStatusCompleted).Count(&summary.CompletedRecruitments)
	case domain.ForeignAgent:
		db := h.selectionRepository.DB().Model(&domain.Selection{})
		db.Where("selected_by = ?", user.ID).Count(&summary.TotalSelections)
		db.Where("selected_by = ? AND status = ?", user.ID, domain.SelectionApproved).Count(&summary.ApprovedSelections)
		db.Where("selected_by = ? AND status = ?", user.ID, domain.SelectionPending).Count(&summary.ActiveRecruitments)
	}
	return summary
}

func (h *AdminAgencyHandler) buildRecentActivity(user *domain.User) []AdminRecentActivityItem {
	items := make([]AdminRecentActivityItem, 0)
	switch user.Role {
	case domain.EthiopianAgent:
		candidates := make([]*domain.Candidate, 0)
		if err := h.candidateRepository.DB().Where("created_by = ?", user.ID).Order("created_at DESC").Limit(5).Find(&candidates).Error; err == nil {
			for _, candidate := range candidates {
				items = append(items, AdminRecentActivityItem{
					ID:         candidate.ID,
					Type:       "candidate",
					Title:      candidate.FullName,
					Status:     string(candidate.Status),
					OccurredAt: candidate.CreatedAt.UTC().Format(time.RFC3339),
				})
			}
		}
	case domain.ForeignAgent:
		selections := make([]*domain.Selection, 0)
		if err := h.selectionRepository.DB().Where("selected_by = ?", user.ID).Order("created_at DESC").Limit(5).Find(&selections).Error; err == nil {
			for _, selection := range selections {
				items = append(items, AdminRecentActivityItem{
					ID:         selection.ID,
					Type:       "selection",
					Title:      selection.CandidateID,
					Status:     string(selection.Status),
					OccurredAt: selection.CreatedAt.UTC().Format(time.RFC3339),
				})
			}
		}
	}
	return items
}

func (h *AdminAgencyHandler) writeAgencyServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "agency not found"})
	case errors.Is(err, service.ErrAgencyAlreadyReviewed):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "agency already reviewed"})
	case errors.Is(err, service.ErrAgencyInvalidStatus):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid agency status"})
	default:
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func matchesAgencyRoleFilter(filter string, role domain.UserRole) bool {
	switch strings.TrimSpace(filter) {
	case "", "all":
		return true
	case string(domain.EthiopianAgent):
		return role == domain.EthiopianAgent
	case string(domain.ForeignAgent):
		return role == domain.ForeignAgent
	default:
		return true
	}
}

func parseAgencyAccountStatus(value string) (domain.AccountStatus, error) {
	switch strings.TrimSpace(value) {
	case string(domain.AccountStatusActive):
		return domain.AccountStatusActive, nil
	case string(domain.AccountStatusRejected):
		return domain.AccountStatusRejected, nil
	case string(domain.AccountStatusSuspended):
		return domain.AccountStatusSuspended, nil
	default:
		return "", errors.New("invalid status")
	}
}
