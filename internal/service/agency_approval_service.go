package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrAgencyAlreadyReviewed = errors.New("agency has already been reviewed")
	ErrAgencyInvalidStatus   = errors.New("invalid agency status")
)

type AgencyApprovalService struct {
	userRepository     domain.UserRepository
	adminRepository    domain.AdminRepository
	approvalRepository domain.AgencyApprovalRequestRepository
	auditRepository    domain.AuditLogRepository
	emailService       EmailService
	platformSettings   PlatformSettingsReader
}

func NewAgencyApprovalService(
	userRepository domain.UserRepository,
	adminRepository domain.AdminRepository,
	approvalRepository domain.AgencyApprovalRequestRepository,
	auditRepository domain.AuditLogRepository,
	emailService EmailService,
) (*AgencyApprovalService, error) {
	if userRepository == nil {
		return nil, fmt.Errorf("user repository is nil")
	}
	if adminRepository == nil {
		return nil, fmt.Errorf("admin repository is nil")
	}
	if approvalRepository == nil {
		return nil, fmt.Errorf("approval repository is nil")
	}
	if auditRepository == nil {
		return nil, fmt.Errorf("audit repository is nil")
	}
	if emailService == nil {
		return nil, fmt.Errorf("email service is nil")
	}

	return &AgencyApprovalService{
		userRepository:     userRepository,
		adminRepository:    adminRepository,
		approvalRepository: approvalRepository,
		auditRepository:    auditRepository,
		emailService:       emailService,
	}, nil
}

func (s *AgencyApprovalService) SetPlatformSettingsReader(platformSettings PlatformSettingsReader) {
	s.platformSettings = platformSettings
}

func (s *AgencyApprovalService) RegisterPendingAgency(user *domain.User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	if user.Role != domain.EthiopianAgent && user.Role != domain.ForeignAgent {
		return nil
	}

	settings := s.currentPlatformSettings()
	request, err := s.approvalRepository.GetByAgencyID(user.ID)
	if err != nil && !errors.Is(err, repository.ErrAgencyApprovalNotFound) {
		return err
	}
	if settings.AutoApproveAgencies {
		now := time.Now().UTC()
		user.AccountStatus = domain.AccountStatusActive
		user.IsActive = true
		if err := s.userRepository.Update(user); err != nil {
			return err
		}

		if request == nil {
			request = &domain.AgencyApprovalRequest{
				AgencyID:   user.ID,
				Status:     domain.AgencyApprovalApproved,
				ReviewedAt: &now,
			}
			if err := s.approvalRepository.Create(request); err != nil {
				return err
			}
		} else {
			request.Status = domain.AgencyApprovalApproved
			request.ReviewedAt = &now
			request.ReviewedBy = nil
			request.RejectionReason = nil
			request.AdminNotes = nil
			if err := s.approvalRepository.Update(request); err != nil {
				return err
			}
		}

		s.sendAgencyEmail(user.Email, "Your agency account has been approved", settings.AgencyApprovalEmailTemplate, map[string]string{
			"company_name": emptyFallback(user.CompanyName, "your agency"),
			"full_name":    emptyFallback(user.FullName, user.Email),
			"reason":       "",
		}, settings)
		return nil
	}

	if request == nil {
		if err := s.approvalRepository.Create(&domain.AgencyApprovalRequest{
			AgencyID: user.ID,
			Status:   domain.AgencyApprovalPending,
		}); err != nil {
			return err
		}
	}
	if !settings.EmailNotificationsEnabled {
		return nil
	}

	admins, err := s.adminRepository.ListActive()
	if err != nil {
		return err
	}
	for _, admin := range admins {
		if admin == nil || strings.TrimSpace(admin.Email) == "" {
			continue
		}
		subject := "New agency awaiting approval"
		body := fmt.Sprintf(
			"A new %s agency registration is pending review.\n\nCompany: %s\nContact: %s\nEmail: %s",
			user.Role,
			emptyFallback(user.CompanyName, "N/A"),
			emptyFallback(user.FullName, "N/A"),
			user.Email,
		)
		go func(email string) {
			_ = s.emailService.Send(email, subject, body)
		}(admin.Email)
	}
	return nil
}

func (s *AgencyApprovalService) ApproveAgency(adminID, agencyID, ipAddress string) error {
	user, request, err := s.loadAgencyRequest(agencyID)
	if err != nil {
		return err
	}
	if request.Status == domain.AgencyApprovalApproved {
		return ErrAgencyAlreadyReviewed
	}

	now := time.Now().UTC()
	user.AccountStatus = domain.AccountStatusActive
	user.IsActive = true
	if err := s.userRepository.Update(user); err != nil {
		return err
	}

	request.Status = domain.AgencyApprovalApproved
	request.ReviewedBy = &adminID
	request.ReviewedAt = &now
	request.RejectionReason = nil
	request.AdminNotes = nil
	if err := s.approvalRepository.Update(request); err != nil {
		return err
	}

	settings := s.currentPlatformSettings()
	s.sendAgencyEmail(user.Email, "Your agency account has been approved", settings.AgencyApprovalEmailTemplate, map[string]string{
		"company_name": emptyFallback(user.CompanyName, "your agency"),
		"full_name":    emptyFallback(user.FullName, user.Email),
		"reason":       "",
	}, settings)

	return s.createAudit(adminID, "approve_agency", "agency", user.ID, map[string]any{
		"email":        user.Email,
		"company_name": user.CompanyName,
		"role":         user.Role,
	}, ipAddress)
}

func (s *AgencyApprovalService) RejectAgency(adminID, agencyID, rejectionReason, adminNotes, ipAddress string) error {
	user, request, err := s.loadAgencyRequest(agencyID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(rejectionReason) == "" {
		return ErrAgencyInvalidStatus
	}

	now := time.Now().UTC()
	user.AccountStatus = domain.AccountStatusRejected
	user.IsActive = true
	if err := s.userRepository.Update(user); err != nil {
		return err
	}

	reason := strings.TrimSpace(rejectionReason)
	request.Status = domain.AgencyApprovalRejected
	request.ReviewedBy = &adminID
	request.ReviewedAt = &now
	request.RejectionReason = &reason
	if strings.TrimSpace(adminNotes) != "" {
		notes := strings.TrimSpace(adminNotes)
		request.AdminNotes = &notes
	} else {
		request.AdminNotes = nil
	}
	if err := s.approvalRepository.Update(request); err != nil {
		return err
	}

	settings := s.currentPlatformSettings()
	s.sendAgencyEmail(user.Email, "Your agency application was rejected", settings.AgencyRejectionEmailTemplate, map[string]string{
		"company_name": emptyFallback(user.CompanyName, "your agency"),
		"full_name":    emptyFallback(user.FullName, user.Email),
		"reason":       reason,
	}, settings)

	return s.createAudit(adminID, "reject_agency", "agency", user.ID, map[string]any{
		"email":            user.Email,
		"company_name":     user.CompanyName,
		"role":             user.Role,
		"rejection_reason": reason,
		"admin_notes":      strings.TrimSpace(adminNotes),
	}, ipAddress)
}

func (s *AgencyApprovalService) UpdateAgencyStatus(adminID, agencyID string, status domain.AccountStatus, reason, ipAddress string) error {
	user, request, err := s.loadAgencyRequest(agencyID)
	if err != nil {
		return err
	}

	switch status {
	case domain.AccountStatusActive:
		request.Status = domain.AgencyApprovalApproved
		request.RejectionReason = nil
	case domain.AccountStatusRejected:
		request.Status = domain.AgencyApprovalRejected
		if strings.TrimSpace(reason) != "" {
			rejectionReason := strings.TrimSpace(reason)
			request.RejectionReason = &rejectionReason
		}
	case domain.AccountStatusSuspended:
		// Keep approval history; the account is only temporarily disabled.
	default:
		return ErrAgencyInvalidStatus
	}

	now := time.Now().UTC()
	user.AccountStatus = status
	user.IsActive = status != domain.AccountStatusSuspended
	if err := s.userRepository.Update(user); err != nil {
		return err
	}

	request.ReviewedBy = &adminID
	request.ReviewedAt = &now
	if err := s.approvalRepository.Update(request); err != nil {
		return err
	}

	return s.createAudit(adminID, "update_agency_status", "agency", user.ID, map[string]any{
		"email":            user.Email,
		"company_name":     user.CompanyName,
		"account_status":   status,
		"rejection_reason": strings.TrimSpace(reason),
	}, ipAddress)
}

func (s *AgencyApprovalService) loadAgencyRequest(agencyID string) (*domain.User, *domain.AgencyApprovalRequest, error) {
	user, err := s.userRepository.GetByID(strings.TrimSpace(agencyID))
	if err != nil {
		return nil, nil, err
	}
	request, err := s.approvalRepository.GetByAgencyID(user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrAgencyApprovalNotFound) {
			request = &domain.AgencyApprovalRequest{
				AgencyID: user.ID,
				Status:   domain.AgencyApprovalPending,
			}
			if createErr := s.approvalRepository.Create(request); createErr != nil {
				return nil, nil, createErr
			}
			return user, request, nil
		}
		return nil, nil, err
	}
	return user, request, nil
}

func (s *AgencyApprovalService) createAudit(adminID, action, targetType, targetID string, details map[string]any, ipAddress string) error {
	payload, err := json.Marshal(details)
	if err != nil {
		return err
	}
	target := strings.TrimSpace(targetID)
	return s.auditRepository.Create(&domain.AuditLog{
		AdminID:    strings.TrimSpace(adminID),
		Action:     action,
		TargetType: strings.TrimSpace(targetType),
		TargetID:   &target,
		Details:    payload,
		IPAddress:  strings.TrimSpace(ipAddress),
	})
}

func (s *AgencyApprovalService) currentPlatformSettings() *domain.PlatformSettings {
	if s.platformSettings == nil {
		return domain.DefaultPlatformSettings()
	}
	settings, err := s.platformSettings.Get()
	if err != nil || settings == nil {
		return domain.DefaultPlatformSettings()
	}
	return settings
}

func (s *AgencyApprovalService) sendAgencyEmail(to, subject, template string, variables map[string]string, settings *domain.PlatformSettings) {
	if settings == nil || !settings.EmailNotificationsEnabled || strings.TrimSpace(to) == "" {
		return
	}
	_ = s.emailService.Send(to, subject, renderTemplate(template, variables))
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
