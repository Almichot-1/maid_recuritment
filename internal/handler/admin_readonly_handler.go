package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/pkg/utils"
)

type AdminCandidateOverview struct {
	ID          string `json:"id"`
	FullName    string `json:"full_name"`
	Age         *int   `json:"age,omitempty"`
	Status      string `json:"status"`
	AgencyID    string `json:"agency_id"`
	AgencyName  string `json:"agency_name"`
	CompanyName string `json:"company_name"`
	CreatedAt   string `json:"created_at"`
}

type AdminSelectionOverview struct {
	ID              string `json:"id"`
	CandidateID     string `json:"candidate_id"`
	CandidateName   string `json:"candidate_name"`
	EthiopianAgency string `json:"ethiopian_agency"`
	ForeignAgency   string `json:"foreign_agency"`
	Status          string `json:"status"`
	SelectedDate    string `json:"selected_date"`
	ApprovalStatus  string `json:"approval_status"`
}

type AdminAuditLogOverview struct {
	ID         string          `json:"id"`
	AdminID    string          `json:"admin_id"`
	AdminName  string          `json:"admin_name"`
	Action     string          `json:"action"`
	TargetType string          `json:"target_type"`
	TargetID   string          `json:"target_id,omitempty"`
	IPAddress  string          `json:"ip_address,omitempty"`
	Details    json.RawMessage `json:"details"`
	CreatedAt  string          `json:"created_at"`
}

type AdminReadonlyHandler struct {
	userRepository      *repository.GormUserRepository
	adminRepository     domain.AdminRepository
	candidateRepository *repository.GormCandidateRepository
	selectionRepository *repository.GormSelectionRepository
	auditRepository     domain.AuditLogRepository
}

func NewAdminReadonlyHandler(
	userRepository *repository.GormUserRepository,
	adminRepository domain.AdminRepository,
	candidateRepository *repository.GormCandidateRepository,
	selectionRepository *repository.GormSelectionRepository,
	auditRepository domain.AuditLogRepository,
) *AdminReadonlyHandler {
	return &AdminReadonlyHandler{
		userRepository:      userRepository,
		adminRepository:     adminRepository,
		candidateRepository: candidateRepository,
		selectionRepository: selectionRepository,
		auditRepository:     auditRepository,
	}
}

func (h *AdminReadonlyHandler) GetCandidates(w http.ResponseWriter, r *http.Request) {
	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	search := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("search")))

	query := h.candidateRepository.DB().Model(&domain.Candidate{})
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}
	if search != "" {
		query = query.Where("LOWER(full_name) LIKE ?", "%"+search+"%")
	}

	candidates := make([]*domain.Candidate, 0)
	if err := query.Order("created_at DESC").Limit(200).Find(&candidates).Error; err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	items := make([]AdminCandidateOverview, 0, len(candidates))
	for _, candidate := range candidates {
		user, err := h.userRepository.GetByID(candidate.CreatedBy)
		if err != nil {
			continue
		}
		items = append(items, AdminCandidateOverview{
			ID:          candidate.ID,
			FullName:    candidate.FullName,
			Age:         candidate.Age,
			Status:      string(candidate.Status),
			AgencyID:    user.ID,
			AgencyName:  user.FullName,
			CompanyName: user.CompanyName,
			CreatedAt:   candidate.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string][]AdminCandidateOverview{"candidates": items})
}

func (h *AdminReadonlyHandler) GetSelections(w http.ResponseWriter, r *http.Request) {
	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	query := h.selectionRepository.DB().Model(&domain.Selection{})
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	selections := make([]*domain.Selection, 0)
	if err := query.Order("created_at DESC").Limit(200).Find(&selections).Error; err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	items := make([]AdminSelectionOverview, 0, len(selections))
	for _, selection := range selections {
		candidate, err := h.candidateRepository.GetByID(selection.CandidateID)
		if err != nil {
			continue
		}
		ethiopianAgency, err := h.userRepository.GetByID(candidate.CreatedBy)
		if err != nil {
			continue
		}
		foreignAgency, err := h.userRepository.GetByID(selection.SelectedBy)
		if err != nil {
			continue
		}
		items = append(items, AdminSelectionOverview{
			ID:              selection.ID,
			CandidateID:     candidate.ID,
			CandidateName:   candidate.FullName,
			EthiopianAgency: emptyFallback(ethiopianAgency.CompanyName, ethiopianAgency.FullName),
			ForeignAgency:   emptyFallback(foreignAgency.CompanyName, foreignAgency.FullName),
			Status:          string(selection.Status),
			SelectedDate:    selection.CreatedAt.UTC().Format(time.RFC3339),
			ApprovalStatus:  string(selection.Status),
		})
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string][]AdminSelectionOverview{"selections": items})
}

func (h *AdminReadonlyHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.auditRepository.List(domain.AuditLogFilters{
		AdminID:    strings.TrimSpace(r.URL.Query().Get("admin_id")),
		Action:     strings.TrimSpace(r.URL.Query().Get("action")),
		TargetType: strings.TrimSpace(r.URL.Query().Get("target_type")),
	})
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	items := make([]AdminAuditLogOverview, 0, len(logs))
	for _, logItem := range logs {
		adminName := ""
		admin, err := h.adminRepository.GetByID(logItem.AdminID)
		if err == nil && admin != nil {
			adminName = admin.FullName
		}
		targetID := ""
		if logItem.TargetID != nil {
			targetID = *logItem.TargetID
		}
		items = append(items, AdminAuditLogOverview{
			ID:         logItem.ID,
			AdminID:    logItem.AdminID,
			AdminName:  adminName,
			Action:     logItem.Action,
			TargetType: logItem.TargetType,
			TargetID:   targetID,
			IPAddress:  logItem.IPAddress,
			Details:    logItem.Details,
			CreatedAt:  logItem.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string][]AdminAuditLogOverview{"logs": items})
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
