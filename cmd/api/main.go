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

	authService, err := service.NewAuthService(userRepository, cfg)
	if err != nil {
		log.Fatalf("failed to initialize auth service: %v", err)
	}
	adminAuthService, err := service.NewAdminAuthService(adminRepository, auditLogRepository, cfg)
	if err != nil {
		log.Fatalf("failed to initialize admin auth service: %v", err)
	}
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

	approvalService, err := service.NewApprovalService(approvalRepository, selectionRepository, candidateRepository, statusStepService, notificationService)
	if err != nil {
		log.Fatalf("failed to initialize approval service: %v", err)
	}
	agencyApprovalService.SetPlatformSettingsReader(platformSettingsService)
	notificationService.SetPlatformSettingsReader(platformSettingsService)
	selectionService.SetPlatformSettingsReader(platformSettingsService)
	approvalService.SetPlatformSettingsReader(platformSettingsService)

	authHandler := handler.NewAuthHandler(authService, userRepository, agencyApprovalService)
	userHandler := handler.NewUserHandler(userRepository)
	dashboardHandler := handler.NewDashboardHandler(candidateRepository, selectionRepository, notificationRepository, pairingService)
	candidateHandler := handler.NewCandidateHandler(candidateService, candidateRepository, selectionRepository, pairingService)
	selectionHandler := handler.NewSelectionHandler(selectionService, candidateRepository, approvalRepository, pairingService)
	approvalHandler := handler.NewApprovalHandler(approvalService, selectionService, candidateRepository, pairingService)
	statusHandler := handler.NewStatusHandler(statusStepService, candidateRepository, selectionRepository, userRepository, pairingService)
	notificationHandler := handler.NewNotificationHandler(notificationRepository)
	pairingHandler := handler.NewPairingHandler(pairingService, userRepository, candidateRepository)
	adminAuthHandler := handler.NewAdminAuthHandler(adminAuthService)
	adminDashboardHandler := handler.NewAdminDashboardHandler(userRepository, candidateRepository, selectionRepository)
	adminAgencyHandler := handler.NewAdminAgencyHandler(userRepository, agencyApprovalRepository, agencyApprovalService, candidateRepository, selectionRepository)
	adminReadonlyHandler := handler.NewAdminReadonlyHandler(userRepository, adminRepository, candidateRepository, selectionRepository, auditLogRepository)
	adminManagementHandler := handler.NewAdminManagementHandler(adminRepository, auditLogRepository, adminAuthService, emailService)
	adminSettingsHandler := handler.NewAdminSettingsHandler(platformSettingsService)
	adminPairingHandler := handler.NewAdminPairingHandler(pairingService, userRepository)
	notificationService.SetRealtimeNotifier(notificationHandler)

	if cfg.RunExpiryScheduler {
		if _, err := jobs.StartExpiryScheduler(selectionService); err != nil {
			log.Fatalf("failed to start expiry scheduler: %v", err)
		}
	} else {
		log.Printf("expiry scheduler disabled for this process")
	}

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.Recoverer)
	router.Use(appmiddleware.CORS(cfg.CORSAllowedOrigins))
	router.Use(appmiddleware.Logging)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/health", handler.Health)
	apiRouter.Route("/auth", func(r chi.Router) {
		r.Use(appmiddleware.PlatformMaintenance(platformSettingsService))
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Group(func(protected chi.Router) {
			protected.Use(appmiddleware.AuthMiddleware(authService))
			protected.Get("/me", authHandler.Me)
		})
	})

	apiRouter.Route("/admin", func(adminRouter chi.Router) {
		adminRouter.Post("/login", adminAuthHandler.Login)

		adminRouter.Group(func(protected chi.Router) {
			protected.Use(appmiddleware.AdminAuthMiddleware(adminAuthService))

			protected.Post("/logout", adminAuthHandler.Logout)
			protected.Get("/analytics/dashboard", adminDashboardHandler.GetStats)
			protected.Get("/audit-logs", adminReadonlyHandler.GetAuditLogs)
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
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Put("/{id}", candidateHandler.UpdateCandidate)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Delete("/{id}", candidateHandler.DeleteCandidate)
			cr.Get("/{id}", candidateHandler.GetCandidate)
			cr.Get("/", candidateHandler.ListCandidates)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Get("/{id}/shares", pairingHandler.GetCandidateShares)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Post("/{id}/publish", candidateHandler.PublishCandidate)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Post("/{id}/documents", candidateHandler.UploadCandidateDocument)
			cr.With(appmiddleware.RequireRole(string(domain.EthiopianAgent))).Post("/{id}/generate-cv", candidateHandler.GenerateCV)
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
			nr.Get("/", notificationHandler.GetNotifications)
			nr.Patch("/{id}/read", notificationHandler.MarkAsRead)
			nr.Post("/mark-all-read", notificationHandler.MarkAllAsRead)
		})

		protected.Route("/dashboard", func(dr chi.Router) {
			dr.Get("/stats", dashboardHandler.GetStats)
		})

		protected.Route("/users", func(ur chi.Router) {
			ur.Patch("/profile", userHandler.UpdateProfile)
			ur.Post("/change-password", userHandler.ChangePassword)
		})
	})

	router.Mount("/api/v1", apiRouter)
	router.Mount("/", apiRouter)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API server listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
