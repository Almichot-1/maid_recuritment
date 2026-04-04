package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/handler"
	"maid-recruitment-tracking/internal/jobs"
	appmiddleware "maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	userRepository, err := repository.NewGormUserRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize user repository: %v", err)
	}
	adminRepository, err := repository.NewGormAdminRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize admin repository: %v", err)
	}
	adminSetupTokenRepository, err := repository.NewGormAdminSetupTokenRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize admin setup token repository: %v", err)
	}
	platformSettingsRepository, err := repository.NewGormPlatformSettingsRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize platform settings repository: %v", err)
	}
	auditLogRepository, err := repository.NewGormAuditLogRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize audit log repository: %v", err)
	}
	agencyApprovalRepository, err := repository.NewGormAgencyApprovalRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize agency approval repository: %v", err)
	}
	agencyPairingRepository, err := repository.NewGormAgencyPairingRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize agency pairing repository: %v", err)
	}
	candidatePairShareRepository, err := repository.NewGormCandidatePairShareRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize candidate pair share repository: %v", err)
	}
	passwordResetRepository, err := repository.NewGormPasswordResetRequestRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize password reset repository: %v", err)
	}
	emailVerificationRepository, err := repository.NewGormEmailVerificationTokenRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize email verification repository: %v", err)
	}
	userSessionRepository, err := repository.NewGormUserSessionRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize user session repository: %v", err)
	}

	authService, err := service.NewAuthService(userRepository, cfg)
	if err != nil {
		log.Fatalf("failed to initialize auth service: %v", err)
	}
	adminAuthService, err := service.NewAdminAuthService(adminRepository, auditLogRepository, cfg)
	if err != nil {
		log.Fatalf("failed to initialize admin auth service: %v", err)
	}
	adminAuthService.SetSetupTokenRepository(adminSetupTokenRepository)
	platformSettingsService, err := service.NewPlatformSettingsService(platformSettingsRepository, auditLogRepository)
	if err != nil {
		log.Fatalf("failed to initialize platform settings service: %v", err)
	}

	candidateRepository, err := repository.NewGormCandidateRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize candidate repository: %v", err)
	}

	documentRepository, err := repository.NewGormDocumentRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize document repository: %v", err)
	}
	passportRepository, err := repository.NewGormPassportDataRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize passport repository: %v", err)
	}
	medicalRepository, err := repository.NewGormMedicalDataRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize medical repository: %v", err)
	}

	selectionRepository, err := repository.NewGormSelectionRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize selection repository: %v", err)
	}

	approvalRepository, err := repository.NewGormApprovalRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize approval repository: %v", err)
	}

	notificationRepository, err := repository.NewGormNotificationRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize notification repository: %v", err)
	}

	statusStepRepository, err := repository.NewGormStatusStepRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize status step repository: %v", err)
	}

	storageService, err := service.NewS3StorageService(cfg)
	if err != nil {
		log.Fatalf("failed to initialize storage service: %v", err)
	}

	pdfService := service.NewPDFService()
	emailService, err := service.NewSMTPEmailService(cfg)
	if err != nil {
		log.Fatalf("failed to initialize email service: %v", err)
	}
	authService.SetEmailService(emailService)
	authService.SetSessionRepository(userSessionRepository)
	authService.SetPasswordResetRepository(passwordResetRepository)
	authService.SetEmailVerificationRepository(emailVerificationRepository)
	agencyApprovalService, err := service.NewAgencyApprovalService(userRepository, adminRepository, agencyApprovalRepository, auditLogRepository, emailService)
	if err != nil {
		log.Fatalf("failed to initialize agency approval service: %v", err)
	}
	pairingService, err := service.NewPairingService(userRepository, agencyPairingRepository, candidatePairShareRepository, selectionRepository, auditLogRepository)
	if err != nil {
		log.Fatalf("failed to initialize pairing service: %v", err)
	}
	notificationService, err := service.NewNotificationService(cfg, notificationRepository, emailService, userRepository, candidateRepository, selectionRepository)
	if err != nil {
		log.Fatalf("failed to initialize notification service: %v", err)
	}

	candidateService, err := service.NewCandidateService(candidateRepository, documentRepository, storageService, pdfService)
	if err != nil {
		log.Fatalf("failed to initialize candidate service: %v", err)
	}
	candidateService.SetUserRepository(userRepository)
	candidateService.SetPassportRepository(passportRepository)
	candidateService.SetMedicalDataRepository(medicalRepository)

	passportOCRService, err := service.NewPassportOCRService(cfg, candidateRepository, passportRepository)
	if err != nil {
		log.Fatalf("failed to initialize passport OCR service: %v", err)
	}
	candidateService.SetPassportOCRService(passportOCRService)
	medicalDocumentService, err := service.NewMedicalDocumentService(cfg, medicalRepository)
	if err != nil {
		log.Fatalf("failed to initialize medical document service: %v", err)
	}
	candidateService.SetMedicalDocumentService(medicalDocumentService)
	expiryWarningJob, err := jobs.NewExpiryWarningJob(selectionRepository, candidateRepository, passportRepository, medicalRepository, notificationService)
	if err != nil {
		log.Fatalf("failed to initialize expiry warning job: %v", err)
	}

	selectionService, err := service.NewSelectionService(selectionRepository, candidateRepository, notificationService)
	if err != nil {
		log.Fatalf("failed to initialize selection service: %v", err)
	}
	selectionService.SetStorageService(storageService)
	selectionService.SetPairingService(pairingService)
	candidateService.SetPairingService(pairingService)

	statusStepService, err := service.NewStatusStepService(statusStepRepository, candidateRepository, selectionRepository, notificationService)
	if err != nil {
		log.Fatalf("failed to initialize status step service: %v", err)
	}
	statusStepService.SetDocumentRepository(documentRepository)
	candidateService.SetStatusStepService(statusStepService)

	approvalService, err := service.NewApprovalService(approvalRepository, selectionRepository, candidateRepository, statusStepService, notificationService)
	if err != nil {
		log.Fatalf("failed to initialize approval service: %v", err)
	}
	agencyApprovalService.SetPlatformSettingsReader(platformSettingsService)
	notificationService.SetPlatformSettingsReader(platformSettingsService)
	selectionService.SetPlatformSettingsReader(platformSettingsService)
	approvalService.SetPlatformSettingsReader(platformSettingsService)

	authHandler := handler.NewAuthHandler(authService, userRepository, agencyApprovalService)
	userHandler := handler.NewUserHandler(userRepository, userSessionRepository, storageService, pairingService)
	dashboardHandler := handler.NewDashboardHandler(candidateRepository, selectionRepository, notificationRepository, pairingService, passportRepository, medicalRepository, statusStepRepository)
	candidateHandler := handler.NewCandidateHandler(candidateService, passportOCRService, candidateRepository, selectionRepository, pairingService)
	candidateHandler.SetDocumentStorage(storageService)
	selectionHandler := handler.NewSelectionHandler(selectionService, candidateRepository, approvalRepository, pairingService)
	selectionHandler.SetDocumentStorage(storageService)
	approvalHandler := handler.NewApprovalHandler(approvalService, selectionService, candidateRepository, pairingService)
	statusHandler := handler.NewStatusHandler(statusStepService, candidateRepository, selectionRepository, documentRepository, userRepository, pairingService)
	notificationHandler := handler.NewNotificationHandler(notificationRepository, cfg.CORSAllowedOrigins)
	pairingHandler := handler.NewPairingHandler(pairingService, userRepository, candidateRepository)
	adminAuthHandler := handler.NewAdminAuthHandler(adminAuthService, adminRepository)
	adminDashboardHandler := handler.NewAdminDashboardHandler(userRepository, candidateRepository, selectionRepository)
	adminAgencyHandler := handler.NewAdminAgencyHandler(userRepository, agencyApprovalRepository, agencyApprovalService, candidateRepository, selectionRepository)
	adminReadonlyHandler := handler.NewAdminReadonlyHandler(userRepository, userSessionRepository, adminRepository, candidateRepository, selectionRepository, auditLogRepository)
	adminManagementHandler := handler.NewAdminManagementHandler(adminRepository, auditLogRepository, adminAuthService, emailService)
	adminSettingsHandler := handler.NewAdminSettingsHandler(platformSettingsService)
	adminPairingHandler := handler.NewAdminPairingHandler(pairingService, userRepository)
	notificationService.SetRealtimeNotifier(notificationHandler)

	if cfg.RunExpiryScheduler {
		if _, err := jobs.StartExpiryScheduler(selectionService, expiryWarningJob); err != nil {
			log.Fatalf("failed to start expiry scheduler: %v", err)
		}
	} else {
		log.Printf("expiry scheduler disabled for this process")
	}

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.Recoverer)
	router.Use(appmiddleware.SecurityHeaders)
	router.Use(appmiddleware.CORS(cfg.CORSAllowedOrigins))
	router.Use(appmiddleware.Logging)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/health", handler.Health)
	apiRouter.Route("/auth", func(r chi.Router) {
		r.Use(appmiddleware.PlatformMaintenance(platformSettingsService))
		r.With(appmiddleware.NewIPRateLimitMiddleware("auth-register", 5, 10*time.Minute)).Post("/register", authHandler.Register)
		r.With(appmiddleware.NewIPRateLimitMiddleware("auth-login", 10, 10*time.Minute)).Post("/login", authHandler.Login)
		r.With(appmiddleware.NewIPRateLimitMiddleware("auth-forgot-password", 5, 10*time.Minute)).Post("/forgot-password/request", authHandler.RequestPasswordReset)
		r.With(appmiddleware.NewIPRateLimitMiddleware("auth-reset-password", 10, 10*time.Minute)).Post("/forgot-password/reset", authHandler.ResetPassword)
		r.With(appmiddleware.NewIPRateLimitMiddleware("auth-verify-email", 20, 10*time.Minute)).Post("/verify-email", authHandler.VerifyEmail)
		r.With(appmiddleware.NewIPRateLimitMiddleware("auth-resend-verification", 5, 10*time.Minute)).Post("/resend-verification", authHandler.ResendVerification)

		r.Group(func(protected chi.Router) {
			protected.Use(appmiddleware.AuthMiddleware(authService))
			protected.Get("/me", authHandler.Me)
			protected.Post("/logout", authHandler.Logout)
		})
	})

	apiRouter.Route("/admin", func(adminRouter chi.Router) {
		adminRouter.With(appmiddleware.NewIPRateLimitMiddleware("admin-login", 8, 10*time.Minute)).Post("/login", adminAuthHandler.Login)
		adminRouter.With(appmiddleware.NewIPRateLimitMiddleware("admin-setup-preview", 10, 10*time.Minute)).Post("/setup/preview", adminAuthHandler.PreviewSetup)
		adminRouter.With(appmiddleware.NewIPRateLimitMiddleware("admin-setup-complete", 10, 10*time.Minute)).Post("/setup/complete", adminAuthHandler.CompleteSetup)

		adminRouter.Group(func(protected chi.Router) {
			protected.Use(appmiddleware.AdminAuthMiddleware(adminAuthService))

			protected.Get("/me", adminAuthHandler.Me)
			protected.Post("/logout", adminAuthHandler.Logout)
			protected.Post("/change-password", adminAuthHandler.ChangePassword)
			protected.Get("/analytics/dashboard", adminDashboardHandler.GetStats)
			protected.Get("/audit-logs", adminReadonlyHandler.GetAuditLogs)
			protected.Get("/agency-logins", adminReadonlyHandler.GetAgencyLogins)
			protected.Get("/candidates", adminReadonlyHandler.GetCandidates)
			protected.Get("/selections", adminReadonlyHandler.GetSelections)
			protected.With(appmiddleware.RequireAdminRole(domain.SuperAdmin)).Get("/settings", adminSettingsHandler.GetSettings)
			protected.With(appmiddleware.RequireAdminRole(domain.SuperAdmin)).Patch("/settings", adminSettingsHandler.UpdateSettings)
			protected.With(appmiddleware.RequireAdminRole(domain.SuperAdmin)).Route("/admins", func(adminsRouter chi.Router) {
				adminsRouter.Get("/", adminManagementHandler.ListAdmins)
				adminsRouter.Post("/", adminManagementHandler.CreateAdmin)
				adminsRouter.Patch("/{id}", adminManagementHandler.UpdateAdmin)
			})

			protected.Route("/agencies", func(agencyRouter chi.Router) {
				agencyRouter.Get("/pending", adminAgencyHandler.GetPendingAgencies)
				agencyRouter.Get("/", adminAgencyHandler.GetAgencies)
				agencyRouter.Get("/{id}", adminAgencyHandler.GetAgency)
				agencyRouter.Get("/{id}/pairings", adminPairingHandler.GetAgencyPairings)
				agencyRouter.Post("/{id}/approve", adminAgencyHandler.ApproveAgency)
				agencyRouter.Post("/{id}/reject", adminAgencyHandler.RejectAgency)
				agencyRouter.Patch("/{id}/status", adminAgencyHandler.UpdateAgencyStatus)
			})
			protected.Route("/pairings", func(pairingRouter chi.Router) {
				pairingRouter.Get("/", adminPairingHandler.GetPairings)
				pairingRouter.Post("/", adminPairingHandler.CreatePairing)
				pairingRouter.Patch("/{id}", adminPairingHandler.UpdatePairing)
			})
		})
	})

	apiRouter.Group(func(wsProtected chi.Router) {
		wsProtected.Use(appmiddleware.PlatformMaintenance(platformSettingsService))
		wsProtected.Use(appmiddleware.AuthMiddleware(authService))
		wsProtected.Use(appmiddleware.NewIPRateLimitMiddleware("notifications-websocket", 20, time.Minute))
		wsProtected.Get("/ws/notifications", notificationHandler.NotificationsWebSocket)
	})

	apiRouter.Group(func(protected chi.Router) {
		protected.Use(appmiddleware.PlatformMaintenance(platformSettingsService))
		protected.Use(appmiddleware.AuthMiddleware(authService))
		protected.Use(appmiddleware.PairingContext)

		protected.Route("/pairings", func(pr chi.Router) {
			pr.Get("/me", pairingHandler.GetMyPairingContext)
			pr.Post("/{pairingId}/candidates/{candidateId}/share", pairingHandler.ShareCandidate)
			pr.Delete("/{pairingId}/candidates/{candidateId}/share", pairingHandler.UnshareCandidate)
		})

		protected.Route("/candidates", func(cr chi.Router) {
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Post("/", candidateHandler.CreateCandidate)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent)), appmiddleware.NewIPRateLimitMiddleware("passport-parse-preview", 12, time.Minute)).Post("/passport/parse-preview", candidateHandler.ParsePassportPreview)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Put("/{id}", candidateHandler.UpdateCandidate)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Delete("/{id}", candidateHandler.DeleteCandidate)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Delete("/{id}/documents/{documentId}", candidateHandler.DeleteCandidateDocument)
			cr.Get("/{id}", candidateHandler.GetCandidate)
			cr.Get("/", candidateHandler.ListCandidates)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Get("/{id}/shares", pairingHandler.GetCandidateShares)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Post("/{id}/publish", candidateHandler.PublishCandidate)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Post("/{id}/documents", candidateHandler.UploadCandidateDocument)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Post("/{id}/passport/parse", candidateHandler.ParsePassport)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Get("/{id}/passport", candidateHandler.GetPassport)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent)), appmiddleware.NewIPRateLimitMiddleware("candidate-generate-cv", 8, time.Minute)).Post("/{id}/generate-cv", candidateHandler.GenerateCV)
			cr.Get("/{id}/download-cv", candidateHandler.DownloadCV)
			cr.With(appmiddleware.RequireRole(string(domain.ForeignAgent))).Post("/{id}/select", selectionHandler.SelectCandidate)
			cr.Get("/{id}/status-steps", statusHandler.GetCandidateStatusSteps)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Patch("/{id}/status-steps/{step_name}", statusHandler.UpdateStatusStep)
		})

		protected.Route("/selections", func(sr chi.Router) {
			sr.Get("/my", selectionHandler.GetMySelections)
			sr.Get("/{id}", selectionHandler.GetSelection)
			sr.With(appmiddleware.RequireRole(string(domain.ForeignAgent))).Post("/{id}/documents", selectionHandler.UploadSelectionDocument)
			sr.Post("/{id}/approve", approvalHandler.ApproveSelection)
			sr.Post("/{id}/reject", approvalHandler.RejectSelection)
			sr.Get("/{id}/approvals", approvalHandler.GetApprovals)
		})

		protected.Route("/notifications", func(nr chi.Router) {
			nr.Get("/summary", notificationHandler.GetSummary)
			nr.Get("/", notificationHandler.GetNotifications)
			nr.Patch("/{id}/read", notificationHandler.MarkAsRead)
			nr.Post("/mark-all-read", notificationHandler.MarkAllAsRead)
		})

		protected.Route("/dashboard", func(dr chi.Router) {
			dr.Get("/home", dashboardHandler.GetHome)
			dr.Get("/stats", dashboardHandler.GetStats)
			dr.Get("/smart-alerts", dashboardHandler.GetSmartAlerts)
		})

		protected.Route("/users", func(ur chi.Router) {
			ur.Patch("/profile", userHandler.UpdateProfile)
			ur.Patch("/sharing-preferences", userHandler.UpdateSharingPreferences)
			ur.Post("/avatar", userHandler.UploadAvatar)
			ur.Delete("/avatar", userHandler.DeleteAvatar)
			ur.Post("/change-password", userHandler.ChangePassword)
			ur.Get("/sessions", userHandler.ListSessions)
			ur.Delete("/sessions/{id}", userHandler.RevokeSession)
			ur.Post("/sessions/logout-all", userHandler.LogoutAllSessions)
		})
	})

	router.Mount("/api/v1", apiRouter)
	router.Mount("/", apiRouter)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	log.Printf("API server listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
