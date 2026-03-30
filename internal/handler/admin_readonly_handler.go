package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
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

type AdminAgencyLoginOverview struct {
	SessionID   string `json:"session_id"`
	UserID      string `json:"user_id"`
	AgencyName  string `json:"agency_name"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	DeviceLabel string `json:"device_label"`
	BrowserName string `json:"browser_name"`
	OSName      string `json:"os_name"`
	IPAddress   string `json:"ip_address,omitempty"`
	LoggedInAt  string `json:"logged_in_at"`
	LastSeenAt  string `json:"last_seen_at"`
	IsActive    bool   `json:"is_active"`
}

type AdminAgencyLoginSummary struct {
	TotalLoginEvents     int64 `json:"total_login_events"`
	ActiveSessions       int64 `json:"active_sessions"`
	EthiopianLoginEvents int64 `json:"ethiopian_login_events"`
	ForeignLoginEvents   int64 `json:"foreign_login_events"`
}

type AdminReadonlyHandler struct {
	userRepository        *repository.GormUserRepository
	userSessionRepository *repository.GormUserSessionRepository
	adminRepository       domain.AdminRepository
	candidateRepository   *repository.GormCandidateRepository
	selectionRepository   *repository.GormSelectionRepository
	auditRepository       domain.AuditLogRepository
}

func NewAdminReadonlyHandler(
	userRepository *repository.GormUserRepository,
	userSessionRepository *repository.GormUserSessionRepository,
	adminRepository domain.AdminRepository,
	candidateRepository *repository.GormCandidateRepository,
	selectionRepository *repository.GormSelectionRepository,
	auditRepository domain.AuditLogRepository,
) *AdminReadonlyHandler {
	return &AdminReadonlyHandler{
		userRepository:        userRepository,
		userSessionRepository: userSessionRepository,
		adminRepository:       adminRepository,
		candidateRepository:   candidateRepository,
		selectionRepository:   selectionRepository,
		auditRepository:       auditRepository,
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

func (h *AdminReadonlyHandler) GetAgencyLogins(w http.ResponseWriter, r *http.Request) {
	type agencyLoginRow struct {
		SessionID   string     `gorm:"column:session_id"`
		UserID      string     `gorm:"column:user_id"`
		UserAgent   string     `gorm:"column:user_agent"`
		IPAddress   string     `gorm:"column:ip_address"`
		LastSeenAt  time.Time  `gorm:"column:last_seen_at"`
		ExpiresAt   time.Time  `gorm:"column:expires_at"`
		RevokedAt   *time.Time `gorm:"column:revoked_at"`
		CreatedAt   time.Time  `gorm:"column:created_at"`
		FullName    string     `gorm:"column:full_name"`
		Email       string     `gorm:"column:email"`
		CompanyName string     `gorm:"column:company_name"`
		Role        string     `gorm:"column:role"`
	}

	if h.userSessionRepository == nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "agency session repository is unavailable"})
		return
	}

	roleFilter := strings.TrimSpace(r.URL.Query().Get("role"))
	search := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("search")))
	now := time.Now().UTC()

	buildQuery := func() *gorm.DB {
		query := h.userSessionRepository.DB().
			Table("user_sessions AS sessions").
			Joins("JOIN users ON users.id = sessions.user_id").
			Where("users.role IN ?", []domain.UserRole{domain.EthiopianAgent, domain.ForeignAgent})

		if roleFilter == string(domain.EthiopianAgent) || roleFilter == string(domain.ForeignAgent) {
			query = query.Where("users.role = ?", roleFilter)
		}
		if search != "" {
			needle := "%" + search + "%"
			query = query.Where(
				"LOWER(COALESCE(users.company_name, '')) LIKE ? OR LOWER(COALESCE(users.full_name, '')) LIKE ? OR LOWER(COALESCE(users.email, '')) LIKE ?",
				needle, needle, needle,
			)
		}
		return query
	}

	var summary AdminAgencyLoginSummary
	buildQuery().Count(&summary.TotalLoginEvents)
	buildQuery().
		Where("sessions.revoked_at IS NULL AND sessions.expires_at > ?", now).
		Count(&summary.ActiveSessions)
	buildQuery().Where("users.role = ?", domain.EthiopianAgent).Count(&summary.EthiopianLoginEvents)
	buildQuery().Where("users.role = ?", domain.ForeignAgent).Count(&summary.ForeignLoginEvents)

	rows := make([]agencyLoginRow, 0)
	if err := buildQuery().
		Select(`
			sessions.id AS session_id,
			sessions.user_id,
			sessions.user_agent,
			sessions.ip_address,
			sessions.last_seen_at,
			sessions.expires_at,
			sessions.revoked_at,
			sessions.created_at,
			users.full_name,
			users.email,
			users.company_name,
			users.role
		`).
		Order("sessions.created_at DESC").
		Limit(200).
		Scan(&rows).Error; err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	items := make([]AdminAgencyLoginOverview, 0, len(rows))
	for _, row := range rows {
		browserName, osName := parseSessionUserAgent(row.UserAgent)
		items = append(items, AdminAgencyLoginOverview{
			SessionID:   row.SessionID,
			UserID:      row.UserID,
			AgencyName:  emptyFallback(row.CompanyName, row.FullName),
			ContactName: row.FullName,
			Email:       row.Email,
			Role:        row.Role,
			DeviceLabel: buildDeviceLabel(browserName, osName),
			BrowserName: browserName,
			OSName:      osName,
			IPAddress:   strings.TrimSpace(row.IPAddress),
			LoggedInAt:  row.CreatedAt.UTC().Format(time.RFC3339),
			LastSeenAt:  row.LastSeenAt.UTC().Format(time.RFC3339),
			IsActive:    row.RevokedAt == nil && row.ExpiresAt.After(now),
		})
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"logins":  items,
	})
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
