package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrSelectionNotPending = errors.New("selection is not pending")
	ErrNotAuthorized       = errors.New("not authorized")
	ErrAlreadyDecided      = errors.New("already decided")
)

type ApprovalService struct {
	approvalRepository  domain.ApprovalRepository
	selectionRepository domain.SelectionRepository
	candidateRepository domain.CandidateRepository
	progressService     *SelectionProgressService
	notificationService NotificationSender
	platformSettings    PlatformSettingsReader
	selectionUpdates    SelectionUpdateSender
	db                  *gorm.DB
}

func NewApprovalService(
	approvalRepository domain.ApprovalRepository,
	selectionRepository domain.SelectionRepository,
	candidateRepository domain.CandidateRepository,
	notificationService NotificationSender,
) (*ApprovalService, error) {
	if approvalRepository == nil {
		return nil, fmt.Errorf("approval repository is nil")
	}
	if selectionRepository == nil {
		return nil, fmt.Errorf("selection repository is nil")
	}
	if candidateRepository == nil {
		return nil, fmt.Errorf("candidate repository is nil")
	}
	if notificationService == nil {
		return nil, fmt.Errorf("notification service is nil")
	}

	dbSource, ok := approvalRepository.(dbProvider)
	if !ok || dbSource.DB() == nil {
		return nil, fmt.Errorf("approval repository does not expose transaction db")
	}

	return &ApprovalService{
		approvalRepository:  approvalRepository,
		selectionRepository: selectionRepository,
		candidateRepository: candidateRepository,
		notificationService: notificationService,
		db:                  dbSource.DB(),
	}, nil
}

func (s *ApprovalService) SetPlatformSettingsReader(platformSettings PlatformSettingsReader) {
	s.platformSettings = platformSettings
}

func (s *ApprovalService) SetSelectionUpdateSender(sender SelectionUpdateSender) {
	s.selectionUpdates = sender
}

func (s *ApprovalService) SetProgressService(progressService *SelectionProgressService) {
	s.progressService = progressService
}

func (s *ApprovalService) ApproveSelection(selectionID, userID string) error {
	selectionID = strings.TrimSpace(selectionID)
	userID = strings.TrimSpace(userID)
	if selectionID == "" {
		return fmt.Errorf("selection id is required")
	}
	if userID == "" {
		return fmt.Errorf("user id is required")
	}

	var (
		finalStatus string
		pairingID   string
	)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		selection, candidate, err := s.loadSelectionAndCandidateForDecision(tx, selectionID)
		if err != nil {
			return err
		}

		pairingID = selection.PairingID

		if !isInvolvedUser(selection, candidate, userID) {
			return ErrNotAuthorized
		}
		existingApproval, err := s.findApprovalInTx(tx, selectionID, userID)
		if err != nil {
			return err
		}
		if existingApproval != nil {
			if existingApproval.Decision == domain.ApprovalApproved {
				return nil
			}
			return ErrAlreadyDecided
		}

		now := time.Now().UTC()
		approval := &domain.Approval{
			ID:          uuid.NewString(),
			SelectionID: selectionID,
			UserID:      userID,
			Decision:    domain.ApprovalApproved,
			DecidedAt:   now,
		}
		if err := tx.Create(approval).Error; err != nil {
			return fmt.Errorf("create approval: %w", err)
		}

		isOwner := strings.TrimSpace(userID) == strings.TrimSpace(candidate.CreatedBy)

		if isOwner {
			if err := tx.Model(&domain.Selection{}).Where("id = ?", selectionID).Update("status", domain.SelectionApproved).Error; err != nil {
				return fmt.Errorf("update selection status approved: %w", err)
			}
			finalStatus = string(domain.SelectionApproved)

			if err := tx.Model(&domain.Candidate{}).Where("id = ?", selection.CandidateID).Updates(map[string]any{
				"status":          domain.CandidateStatusInProgress,
				"locked_by":       nil,
				"locked_at":       nil,
				"lock_expires_at": nil,
			}).Error; err != nil {
				return fmt.Errorf("update candidate in_progress: %w", err)
			}

			// Create progress tracking record inside the transaction
			progress := &domain.SelectionProgress{
				ID:            uuid.NewString(),
				SelectionID:   selectionID,
				UpdatedBy:     candidate.CreatedBy,
				COCStatus:     domain.ProgressStatusPending,
				MedicalStatus: domain.ProgressStatusPending,
				VisaStatus:    domain.VisaStatusPending,
				TicketStatus:  domain.TicketStatusPending,
				ArrivalStatus: domain.ArrivalStatusNotArrived,
			}
			if err := tx.Create(progress).Error; err != nil {
				return fmt.Errorf("create progress tracking: %w", err)
			}

			if err := s.notificationService.Send(candidate.CreatedBy, "Selection approved", "The selection has been approved and recruitment tracking has started.", "approval", "selection", selectionID); err != nil {
				return fmt.Errorf("notify owner approved: %w", err)
			}
			if err := s.notificationService.Send(selection.SelectedBy, "Selection approved", "The Ethiopian agency approved your selection. Recruitment tracking has started.", "approval", "selection", selectionID); err != nil {
				return fmt.Errorf("notify selector approved: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.selectionUpdates != nil && finalStatus != "" {
		s.selectionUpdates.PushSelectionUpdate(selectionID, finalStatus, "approve", pairingID)
	}

	return nil
}

func (s *ApprovalService) RejectSelection(selectionID, userID, reason string) error {
	selectionID = strings.TrimSpace(selectionID)
	userID = strings.TrimSpace(userID)
	reason = strings.TrimSpace(reason)

	if selectionID == "" {
		return fmt.Errorf("selection id is required")
	}
	if userID == "" {
		return fmt.Errorf("user id is required")
	}

	var pairingID string

	err := s.db.Transaction(func(tx *gorm.DB) error {
		selection, candidate, err := s.loadSelectionAndCandidateForDecision(tx, selectionID)
		if err != nil {
			return err
		}

		pairingID = selection.PairingID

		if !isInvolvedUser(selection, candidate, userID) {
			return ErrNotAuthorized
		}

		existingApproval, err := s.findApprovalInTx(tx, selectionID, userID)
		if err != nil {
			return err
		}
		if existingApproval != nil {
			return ErrAlreadyDecided
		}

		approval := &domain.Approval{
			ID:          uuid.NewString(),
			SelectionID: selectionID,
			UserID:      userID,
			Decision:    domain.ApprovalRejected,
			DecidedAt:   time.Now().UTC(),
		}
		if err := tx.Create(approval).Error; err != nil {
			return fmt.Errorf("create rejection approval: %w", err)
		}

		if err := tx.Model(&domain.Selection{}).Where("id = ?", selectionID).Update("status", domain.SelectionRejected).Error; err != nil {
			return fmt.Errorf("update selection rejected: %w", err)
		}

		if err := tx.Model(&domain.Candidate{}).Where("id = ?", selection.CandidateID).Updates(map[string]any{
			"status":          domain.CandidateStatusAvailable,
			"locked_by":       nil,
			"locked_at":       nil,
			"lock_expires_at": nil,
		}).Error; err != nil {
			return fmt.Errorf("update candidate available: %w", err)
		}

		message := "Selection was rejected."
		if reason != "" {
			message = fmt.Sprintf("Selection was rejected. Reason: %s", reason)
		}

		if err := s.notificationService.Send(candidate.CreatedBy, "Selection rejected", message, "rejection", "selection", selectionID); err != nil {
			return fmt.Errorf("notify owner rejection: %w", err)
		}
		if err := s.notificationService.Send(selection.SelectedBy, "Selection rejected", message, "rejection", "selection", selectionID); err != nil {
			return fmt.Errorf("notify selector rejection: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.selectionUpdates != nil {
		s.selectionUpdates.PushSelectionUpdate(selectionID, string(domain.SelectionRejected), "reject", pairingID)
	}

	return nil
}

func (s *ApprovalService) GetApprovals(selectionID string) ([]*domain.Approval, error) {
	if strings.TrimSpace(selectionID) == "" {
		return nil, fmt.Errorf("selection id is required")
	}
	return s.approvalRepository.GetBySelectionID(selectionID)
}

func (s *ApprovalService) loadSelectionAndCandidateForDecision(tx *gorm.DB, selectionID string) (*domain.Selection, *domain.Candidate, error) {
	var selection domain.Selection
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", selectionID).First(&selection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, repository.ErrSelectionNotFound
		}
		return nil, nil, fmt.Errorf("load selection: %w", err)
	}

	if selection.Status != domain.SelectionPending {
		return nil, nil, ErrSelectionNotPending
	}
	var candidate domain.Candidate
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", selection.CandidateID).First(&candidate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, repository.ErrCandidateNotFound
		}
		return nil, nil, fmt.Errorf("load candidate: %w", err)
	}

	return &selection, &candidate, nil
}

func (s *ApprovalService) findApprovalInTx(tx *gorm.DB, selectionID, userID string) (*domain.Approval, error) {
	var approval domain.Approval
	err := tx.Where("selection_id = ? AND user_id = ?", selectionID, userID).First(&approval).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("load user approval: %w", err)
	}
	return &approval, nil
}

func (s *ApprovalService) getApprovalsInTx(tx *gorm.DB, selectionID string) ([]*domain.Approval, error) {
	approvals := make([]*domain.Approval, 0)
	if err := tx.Where("selection_id = ?", selectionID).Find(&approvals).Error; err != nil {
		return nil, fmt.Errorf("load approvals in tx: %w", err)
	}
	return approvals, nil
}

func isInvolvedUser(selection *domain.Selection, candidate *domain.Candidate, userID string) bool {
	trimmedUserID := strings.TrimSpace(userID)
	return strings.TrimSpace(selection.SelectedBy) == trimmedUserID || strings.TrimSpace(candidate.CreatedBy) == trimmedUserID
}
