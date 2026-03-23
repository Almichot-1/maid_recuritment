package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var ErrStepNotFound = errors.New("step not found")
var ErrInvalidStepTransition = errors.New("invalid step transition")
var ErrStepFailureReasonRequired = errors.New("add a short reason before marking this step as failed")

type StatusStepService struct {
	statusStepRepository domain.StatusStepRepository
	candidateRepository  domain.CandidateRepository
	selectionRepository  domain.SelectionRepository
	notificationService  NotificationSender
	db                   *gorm.DB
}

func NewStatusStepService(
	statusStepRepository domain.StatusStepRepository,
	candidateRepository domain.CandidateRepository,
	selectionRepository domain.SelectionRepository,
	notificationService NotificationSender,
) (*StatusStepService, error) {
	if statusStepRepository == nil {
		return nil, fmt.Errorf("status step repository is nil")
	}
	if candidateRepository == nil {
		return nil, fmt.Errorf("candidate repository is nil")
	}
	if selectionRepository == nil {
		return nil, fmt.Errorf("selection repository is nil")
	}
	if notificationService == nil {
		return nil, fmt.Errorf("notification service is nil")
	}

	dbSource, ok := statusStepRepository.(dbProvider)
	if !ok || dbSource.DB() == nil {
		return nil, fmt.Errorf("status step repository does not expose transaction db")
	}

	return &StatusStepService{
		statusStepRepository: statusStepRepository,
		candidateRepository:  candidateRepository,
		selectionRepository:  selectionRepository,
		notificationService:  notificationService,
		db:                   dbSource.DB(),
	}, nil
}

func (s *StatusStepService) InitializeSteps(candidateID string) error {
	if strings.TrimSpace(candidateID) == "" {
		return fmt.Errorf("candidate id is required")
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.initializeStepsWithTx(tx, candidateID, candidate.CreatedBy)
	})
}

func (s *StatusStepService) initializeStepsWithTx(tx *gorm.DB, candidateID, ownerID string) error {
	existingSteps := make([]*domain.StatusStep, 0)
	if err := tx.Where("candidate_id = ?", candidateID).Order("created_at ASC").Find(&existingSteps).Error; err != nil {
		return fmt.Errorf("load existing steps: %w", err)
	}

	if len(existingSteps) == 0 {
		return createPendingStepsWithTx(tx, candidateID, ownerID)
	}

	if matchesCurrentWorkflow(existingSteps) {
		return nil
	}

	return rebuildStepsForCurrentWorkflowWithTx(tx, candidateID, ownerID, existingSteps)
}

func (s *StatusStepService) UpdateStep(candidateID, stepName, updatedBy string, status domain.StepStatus, notes string) error {
	if strings.TrimSpace(candidateID) == "" {
		return fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(stepName) == "" {
		return fmt.Errorf("step name is required")
	}
	if strings.TrimSpace(updatedBy) == "" {
		return fmt.Errorf("updatedBy is required")
	}
	if !isValidStepStatus(status) {
		return fmt.Errorf("invalid step status")
	}
	if status == domain.Failed && strings.TrimSpace(notes) == "" {
		return ErrStepFailureReasonRequired
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(updatedBy) {
		return ErrNotAuthorized
	}
	if candidate.Status != "" && candidate.Status != domain.CandidateStatusInProgress && candidate.Status != domain.CandidateStatusCompleted {
		return ErrInvalidStepTransition
	}
	if s.db != nil {
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			return s.initializeStepsWithTx(tx, candidateID, candidate.CreatedBy)
		}); err != nil {
			return err
		}
	}

	steps, err := s.statusStepRepository.GetByCandidateID(candidateID)
	if err != nil {
		return err
	}

	var target *domain.StatusStep
	targetIndex := -1
	for idx, step := range steps {
		if step == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(step.StepName), strings.TrimSpace(stepName)) {
			target = step
			targetIndex = idx
			break
		}
	}
	if target == nil {
		return ErrStepNotFound
	}

	if !canTransitionStep(target.StepStatus, status) {
		return ErrInvalidStepTransition
	}

	if status == domain.Completed {
		for idx := 0; idx < targetIndex; idx++ {
			if steps[idx] == nil {
				continue
			}
			if steps[idx].StepStatus != domain.Completed {
				return ErrInvalidStepTransition
			}
		}
	}

	target.StepStatus = status
	target.UpdatedBy = strings.TrimSpace(updatedBy)
	target.Notes = strings.TrimSpace(notes)
	if status == domain.Completed {
		now := time.Now().UTC()
		target.CompletedAt = &now
	} else {
		target.CompletedAt = nil
	}

	if err := s.statusStepRepository.Update(target); err != nil {
		if errors.Is(err, repository.ErrStatusStepNotFound) {
			return ErrStepNotFound
		}
		return err
	}

	nextCandidateStatus := domain.CandidateStatusInProgress
	allCompleted := true
	for _, step := range steps {
		if step == nil {
			continue
		}
		if step.StepStatus != domain.Completed {
			allCompleted = false
			break
		}
	}
	if allCompleted {
		nextCandidateStatus = domain.CandidateStatusCompleted
	}

	if s.db != nil {
		if err := s.db.Model(&domain.Candidate{}).
			Where("id = ?", candidateID).
			Update("status", nextCandidateStatus).Error; err != nil {
			return fmt.Errorf("update candidate status after step change: %w", err)
		}
	}

	selection, err := s.selectionRepository.GetByCandidateID(candidateID)
	if err == nil && selection != nil {
		if allCompleted {
			_ = s.notificationService.Send(
				candidate.CreatedBy,
				"Recruitment completed",
				"All recruitment steps have been completed for this candidate.",
				"status_update",
				"candidate",
				candidateID,
			)
			_ = s.notificationService.Send(
				selection.SelectedBy,
				"Recruitment completed",
				"All recruitment steps have been completed for this candidate.",
				"status_update",
				"candidate",
				candidateID,
			)
			return nil
		}

		if status == domain.Failed {
			failureMessage := fmt.Sprintf("Step '%s' was marked as failed.", target.StepName)
			if strings.TrimSpace(target.Notes) != "" {
				failureMessage = fmt.Sprintf("%s Reason: %s", failureMessage, target.Notes)
			}
			_ = s.notificationService.Send(
				selection.SelectedBy,
				"Recruitment issue reported",
				failureMessage,
				"status_update",
				"candidate",
				candidateID,
			)
			return nil
		}

		_ = s.notificationService.Send(
			selection.SelectedBy,
			"Recruitment progress updated",
			fmt.Sprintf("Step '%s' updated to '%s'.", target.StepName, target.StepStatus),
			"status_update",
			"candidate",
			candidateID,
		)
	}

	return nil
}

func (s *StatusStepService) GetCandidateProgress(candidateID string) ([]*domain.StatusStep, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if s.candidateRepository == nil {
		return s.statusStepRepository.GetByCandidateID(candidateID)
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return nil, err
	}
	if candidate.Status != "" && candidate.Status != domain.CandidateStatusInProgress && candidate.Status != domain.CandidateStatusCompleted {
		return s.statusStepRepository.GetByCandidateID(candidateID)
	}
	if s.db != nil {
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			return s.initializeStepsWithTx(tx, candidateID, candidate.CreatedBy)
		}); err != nil {
			return nil, err
		}
	}

	return s.statusStepRepository.GetByCandidateID(candidateID)
}

func predefinedStepNames() []string {
	return []string{
		domain.Medical,
		domain.CoCPending,
		domain.CoCOnline,
		domain.LMISPending,
		domain.LMISIssued,
		domain.TicketPending,
		domain.TicketBooked,
		domain.TicketConfirmed,
		domain.Arrived,
	}
}

func createPendingStepsWithTx(tx *gorm.DB, candidateID, ownerID string) error {
	now := time.Now().UTC()
	for index, stepName := range predefinedStepNames() {
		timestamp := now.Add(time.Duration(index) * time.Millisecond)
		step := &domain.StatusStep{
			ID:          uuid.NewString(),
			CandidateID: candidateID,
			StepName:    stepName,
			StepStatus:  domain.Pending,
			UpdatedBy:   ownerID,
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
		}
		if err := tx.Create(step).Error; err != nil {
			return fmt.Errorf("create initial status step %s: %w", stepName, err)
		}
	}

	return nil
}

func matchesCurrentWorkflow(steps []*domain.StatusStep) bool {
	workflow := predefinedStepNames()
	if len(steps) != len(workflow) {
		return false
	}

	for index, stepName := range workflow {
		if steps[index] == nil || !strings.EqualFold(strings.TrimSpace(steps[index].StepName), strings.TrimSpace(stepName)) {
			return false
		}
	}

	return true
}

func rebuildStepsForCurrentWorkflowWithTx(tx *gorm.DB, candidateID, ownerID string, existingSteps []*domain.StatusStep) error {
	workflow := predefinedStepNames()
	completedLegacy := 0
	hasInProgress := false

	for _, step := range existingSteps {
		if step == nil {
			continue
		}
		switch step.StepStatus {
		case domain.Completed:
			completedLegacy++
		case domain.InProgress:
			hasInProgress = true
		}
	}

	completedTarget := approximateCompletedSteps(completedLegacy, len(existingSteps), len(workflow))
	if completedTarget >= len(workflow) {
		hasInProgress = false
		completedTarget = len(workflow)
	}

	if err := tx.Where("candidate_id = ?", candidateID).Delete(&domain.StatusStep{}).Error; err != nil {
		return fmt.Errorf("reset legacy workflow steps: %w", err)
	}

	now := time.Now().UTC()
	for index, stepName := range workflow {
		timestamp := now.Add(time.Duration(index) * time.Millisecond)
		status := domain.Pending
		var completedAt *time.Time

		if index < completedTarget {
			status = domain.Completed
			completed := timestamp
			completedAt = &completed
		} else if hasInProgress && index == completedTarget {
			status = domain.InProgress
		}

		step := &domain.StatusStep{
			ID:          uuid.NewString(),
			CandidateID: candidateID,
			StepName:    stepName,
			StepStatus:  status,
			CompletedAt: completedAt,
			UpdatedBy:   ownerID,
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
		}
		if err := tx.Create(step).Error; err != nil {
			return fmt.Errorf("create workflow step %s: %w", stepName, err)
		}
	}

	return nil
}

func approximateCompletedSteps(completedLegacy, totalLegacy, totalCurrent int) int {
	if totalLegacy <= 0 || totalCurrent <= 0 {
		return 0
	}

	scaled := math.Round((float64(completedLegacy) / float64(totalLegacy)) * float64(totalCurrent))
	if scaled < 0 {
		return 0
	}
	if scaled > float64(totalCurrent) {
		return totalCurrent
	}
	return int(scaled)
}

func isValidStepStatus(status domain.StepStatus) bool {
	switch status {
	case domain.Pending, domain.InProgress, domain.Completed, domain.Failed:
		return true
	default:
		return false
	}
}

func canTransitionStep(current, next domain.StepStatus) bool {
	if current == domain.Completed {
		return next == domain.Completed
	}

	switch current {
	case domain.Pending:
		return next == domain.InProgress
	case domain.InProgress:
		return next == domain.Completed || next == domain.Failed
	case domain.Failed:
		return next == domain.InProgress || next == domain.Failed
	default:
		return false
	}
}
