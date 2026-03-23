package main

import (
	"log"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
)

type noopAuditLogRepository struct{}

func (noopAuditLogRepository) Create(*domain.AuditLog) error { return nil }
func (noopAuditLogRepository) List(domain.AuditLogFilters) ([]*domain.AuditLog, error) {
	return []*domain.AuditLog{}, nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	userRepository, err := repository.NewGormUserRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize user repository: %v", err)
	}

	candidateRepository, err := repository.NewGormCandidateRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize candidate repository: %v", err)
	}

	selectionRepository, err := repository.NewGormSelectionRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize selection repository: %v", err)
	}

	notificationRepository, err := repository.NewGormNotificationRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize notification repository: %v", err)
	}

	platformSettingsRepository, err := repository.NewGormPlatformSettingsRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize platform settings repository: %v", err)
	}

	emailService, err := service.NewSMTPEmailService(cfg)
	if err != nil {
		log.Fatalf("failed to initialize email service: %v", err)
	}

	notificationService, err := service.NewNotificationService(cfg, notificationRepository, emailService, userRepository, candidateRepository, selectionRepository)
	if err != nil {
		log.Fatalf("failed to initialize notification service: %v", err)
	}

	selectionService, err := service.NewSelectionService(selectionRepository, candidateRepository, notificationService)
	if err != nil {
		log.Fatalf("failed to initialize selection service: %v", err)
	}

	platformSettingsService, err := service.NewPlatformSettingsService(platformSettingsRepository, noopAuditLogRepository{})
	if err != nil {
		log.Fatalf("failed to initialize platform settings service: %v", err)
	}

	selectionService.SetPlatformSettingsReader(platformSettingsService)
	notificationService.SetPlatformSettingsReader(platformSettingsService)

	if err := selectionService.ProcessExpiredSelections(); err != nil {
		log.Fatalf("expiry worker failed: %v", err)
	}

	log.Println("expiry worker completed successfully")
}
