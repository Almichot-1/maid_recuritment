package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

type NotificationService struct {
	notificationRepository domain.NotificationRepository
	emailService           EmailService
	userRepository         domain.UserRepository
	candidateRepository    domain.CandidateRepository
	selectionRepository    domain.SelectionRepository
	platformSettings       PlatformSettingsReader
	baseURL                string
	realtimeNotifier       RealtimeNotifier
}

type RealtimeNotifier interface {
	PushToUser(userID string, notification *domain.Notification)
}

func NewNotificationService(
	cfg *config.Config,
	notificationRepository domain.NotificationRepository,
	emailService EmailService,
	userRepository domain.UserRepository,
	candidateRepository domain.CandidateRepository,
	selectionRepository domain.SelectionRepository,
) (*NotificationService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if notificationRepository == nil {
		return nil, fmt.Errorf("notification repository is nil")
	}
	if emailService == nil {
		return nil, fmt.Errorf("email service is nil")
	}
	if userRepository == nil {
		return nil, fmt.Errorf("user repository is nil")
	}
	if candidateRepository == nil {
		return nil, fmt.Errorf("candidate repository is nil")
	}
	if selectionRepository == nil {
		return nil, fmt.Errorf("selection repository is nil")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.AppBaseURL), "/")

	return &NotificationService{
		notificationRepository: notificationRepository,
		emailService:           emailService,
		userRepository:         userRepository,
		candidateRepository:    candidateRepository,
		selectionRepository:    selectionRepository,
		baseURL:                baseURL,
	}, nil
}

func (s *NotificationService) SetRealtimeNotifier(realtimeNotifier RealtimeNotifier) {
	s.realtimeNotifier = realtimeNotifier
}

func (s *NotificationService) SetPlatformSettingsReader(platformSettings PlatformSettingsReader) {
	s.platformSettings = platformSettings
}

func (s *NotificationService) IsForeignAgent(userID string) (bool, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return false, err
	}
	return user.Role == domain.ForeignAgent, nil
}

func (s *NotificationService) Send(userID, title, message, notificationType, relatedEntityType, relatedEntityID string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user id is required")
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return err
	}

	notification := &domain.Notification{
		ID:                uuid.NewString(),
		UserID:            userID,
		Title:             title,
		Message:           message,
		Type:              domain.NotificationType(notificationType),
		IsRead:            false,
		RelatedEntityType: relatedEntityType,
		CreatedAt:         time.Now().UTC(),
	}
	if strings.TrimSpace(relatedEntityID) != "" {
		relatedID := strings.TrimSpace(relatedEntityID)
		notification.RelatedEntityID = &relatedID
	}

	if err := s.notificationRepository.Create(notification); err != nil {
		return err
	}

	if s.realtimeNotifier != nil {
		s.realtimeNotifier.PushToUser(userID, notification)
	}

	settings := s.currentPlatformSettings()
	if strings.TrimSpace(user.Email) != "" && settings.EmailNotificationsEnabled {
		body := message
		variables := map[string]string{
			"title":          title,
			"message":        message,
			"candidate_name": s.resolveCandidateName(relatedEntityType, relatedEntityID),
			"company_name":   emptyFallback(user.CompanyName, user.Email),
			"full_name":      emptyFallback(user.FullName, user.Email),
		}
		switch strings.TrimSpace(notificationType) {
		case string(domain.NotificationExpiry):
			body = renderTemplate(settings.ExpiryNotificationEmailTemplate, variables)
		default:
			body = renderTemplate(settings.SelectionNotificationEmailTemplate, variables)
		}
		go func(email, subject, body string) {
			if err := s.emailService.Send(email, subject, body); err != nil {
				log.Printf("send notification email failed: %v", err)
			}
		}(user.Email, title, body)
	}

	return nil
}

func (s *NotificationService) NotifySelection(candidateID, selectedBy string) error {
	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return err
	}
	owner, err := s.userRepository.GetByID(candidate.CreatedBy)
	if err != nil {
		return err
	}
	selector, err := s.userRepository.GetByID(selectedBy)
	if err != nil {
		return err
	}

	agency := selector.CompanyName
	if strings.TrimSpace(agency) == "" {
		agency = selector.FullName
	}

	message := fmt.Sprintf("Your candidate %s has been selected by %s", candidate.FullName, agency)
	return s.Send(owner.ID, "Candidate selected", message+"\n"+s.candidateLink(candidate.ID), string(domain.NotificationSelection), "candidate", candidate.ID)
}

func (s *NotificationService) NotifyApproval(selectionID string) error {
	selection, err := s.selectionRepository.GetByID(selectionID)
	if err != nil {
		return err
	}
	candidate, err := s.candidateRepository.GetByID(selection.CandidateID)
	if err != nil {
		return err
	}

	msg := "Selection approved, recruitment process starting"
	if err := s.Send(candidate.CreatedBy, "Selection approved", msg+"\n"+s.selectionLink(selection.ID), string(domain.NotificationApproval), "selection", selection.ID); err != nil {
		return err
	}
	if err := s.Send(selection.SelectedBy, "Selection approved", msg+"\n"+s.selectionLink(selection.ID), string(domain.NotificationApproval), "selection", selection.ID); err != nil {
		return err
	}
	return nil
}

func (s *NotificationService) NotifyRejection(selectionID string, rejectedBy string) error {
	selection, err := s.selectionRepository.GetByID(selectionID)
	if err != nil {
		return err
	}
	candidate, err := s.candidateRepository.GetByID(selection.CandidateID)
	if err != nil {
		return err
	}

	rejectedParty := "a party"
	if strings.TrimSpace(rejectedBy) == strings.TrimSpace(candidate.CreatedBy) {
		rejectedParty = "candidate owner"
	} else if strings.TrimSpace(rejectedBy) == strings.TrimSpace(selection.SelectedBy) {
		rejectedParty = "selector"
	}

	msg := fmt.Sprintf("Selection rejected by %s, candidate now available", rejectedParty)
	if err := s.Send(candidate.CreatedBy, "Selection rejected", msg+"\n"+s.selectionLink(selection.ID), string(domain.NotificationRejection), "selection", selection.ID); err != nil {
		return err
	}
	if err := s.Send(selection.SelectedBy, "Selection rejected", msg+"\n"+s.selectionLink(selection.ID), string(domain.NotificationRejection), "selection", selection.ID); err != nil {
		return err
	}
	return nil
}

func (s *NotificationService) NotifyStatusUpdate(candidateID, stepName string) error {
	selection, err := s.selectionRepository.GetByCandidateID(candidateID)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Progress update: %s", strings.TrimSpace(stepName))
	return s.Send(selection.SelectedBy, "Progress update", message+"\n"+s.candidateLink(candidateID), string(domain.NotificationStatusUpdate), "candidate", candidateID)
}

func (s *NotificationService) NotifyExpiry(selectionID string) error {
	selection, err := s.selectionRepository.GetByID(selectionID)
	if err != nil {
		return err
	}
	candidate, err := s.candidateRepository.GetByID(selection.CandidateID)
	if err != nil {
		return err
	}

	msg := "Selection expired after 24 hours, candidate released"
	if err := s.Send(candidate.CreatedBy, "Selection expired", msg+"\n"+s.selectionLink(selection.ID), string(domain.NotificationExpiry), "selection", selection.ID); err != nil {
		return err
	}
	if err := s.Send(selection.SelectedBy, "Selection expired", msg+"\n"+s.selectionLink(selection.ID), string(domain.NotificationExpiry), "selection", selection.ID); err != nil {
		return err
	}
	return nil
}

func (s *NotificationService) candidateLink(candidateID string) string {
	if s.baseURL == "" {
		return fmt.Sprintf("/candidates/%s", candidateID)
	}
	return fmt.Sprintf("%s/candidates/%s", s.baseURL, candidateID)
}

func (s *NotificationService) selectionLink(selectionID string) string {
	if s.baseURL == "" {
		return fmt.Sprintf("/selections/%s", selectionID)
	}
	return fmt.Sprintf("%s/selections/%s", s.baseURL, selectionID)
}

func (s *NotificationService) currentPlatformSettings() *domain.PlatformSettings {
	if s.platformSettings == nil {
		return domain.DefaultPlatformSettings()
	}
	settings, err := s.platformSettings.Get()
	if err != nil || settings == nil {
		return domain.DefaultPlatformSettings()
	}
	return settings
}

func (s *NotificationService) resolveCandidateName(relatedEntityType, relatedEntityID string) string {
	entityType := strings.TrimSpace(relatedEntityType)
	entityID := strings.TrimSpace(relatedEntityID)
	if entityID == "" {
		return ""
	}

	if entityType == "candidate" {
		candidate, err := s.candidateRepository.GetByID(entityID)
		if err == nil && candidate != nil {
			return candidate.FullName
		}
		return ""
	}

	if entityType == "selection" {
		selection, err := s.selectionRepository.GetByID(entityID)
		if err != nil || selection == nil {
			return ""
		}
		candidate, err := s.candidateRepository.GetByID(selection.CandidateID)
		if err == nil && candidate != nil {
			return candidate.FullName
		}
	}

	return ""
}
