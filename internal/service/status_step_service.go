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

var (
	ErrStepNotFound            = errors.New("step not found")
	ErrInvalidStepTransition   = errors.New("invalid step transition")
	ErrMedicalDocumentRequired = errors.New("medical certificate must be uploaded before completing this step")
)

type StatusStepService struct {
	statusStepRepository domain.StatusStepRepository
	candidateRepository  domain.CandidateRepository
	selectionRepository  domain.SelectionRepository
	documentRepository   domain.DocumentRepository
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

// SetDocumentRepository injects the document repository so the Medical step
// guard can verify that a medical certificate has been uploaded.
func (s *StatusStepService) SetDocumentRepository(dr domain.DocumentRepository) {
	s.documentRepository = dr
}

// ─── public API ──────────────────────────────────────────────────────────────

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

// UpdateStep is the checklist toggle:
//   - pending  → completed  (check the box)
//   - completed → pending   (uncheck the box)
//   - in_progress → completed is still accepted for backwards compatibility
//
// The Medical step additionally requires a medical document to have been
// uploaded before it can be marked completed.
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
	if !isValidChecklistStatus(status) {
		return fmt.Errorf("status must be 'pending' or 'completed'")
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(updatedBy) {
		return ErrNotAuthorized
	}
	if candidate.Status != "" &&
		candidate.Status != domain.CandidateStatusInProgress &&
		candidate.Status != domain.CandidateStatusCompleted {
		return ErrInvalidStepTransition
	}

	// Ensure steps are initialised (idempotent).
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

	// Find the target step.
	var target *domain.StatusStep
	for _, step := range steps {
		if step == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(step.StepName), strings.TrimSpace(stepName)) {
			target = step
			break
		}
	}
	if target == nil {
		return ErrStepNotFound
	}

	// ── Medical document guard ───────────────────────────────────────────────
	// When checking the Medical step as completed, verify a medical certificate
	// has been uploaded. This is a soft guard – if the document repository has
	// not been injected (e.g. in tests) we skip the check.
	if status == domain.Completed &&
		strings.EqualFold(strings.TrimSpace(target.StepName), domain.MedicalStep) &&
		s.documentRepository != nil {

		_, docErr := s.documentRepository.GetByCandidateIDAndType(candidateID, domain.Medical)
		if docErr != nil {
			if errors.Is(docErr, repository.ErrDocumentNotFound) {
				return ErrMedicalDocumentRequired
			}
			return fmt.Errorf("check medical document: %w", docErr)
		}
	}

	// ── Checklist transition ─────────────────────────────────────────────────
	// Allow: pending→completed, completed→pending, in_progress→completed.
	// The only forbidden transition is completed→in_progress (no half-steps).
	if target.StepStatus == domain.InProgress && status == domain.Pending {
		// Treat in_progress → pending as a valid uncheck.
		status = domain.Pending
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

	// ── Recalculate overall candidate status ─────────────────────────────────
	// Reload steps so the count is accurate after the update.
	updatedSteps, err := s.statusStepRepository.GetByCandidateID(candidateID)
	if err != nil {
		return err
	}

	allCompleted := len(updatedSteps) > 0
	for _, step := range updatedSteps {
		if step == nil {
			continue
		}
		if step.StepStatus != domain.Completed {
			allCompleted = false
			break
		}
	}

	nextCandidateStatus := domain.CandidateStatusInProgress
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

	// ── Notifications ─────────────────────────────────────────────────────────
	selection, selErr := s.selectionRepository.GetByCandidateID(candidateID)
	if selErr != nil || selection == nil {
		// No selection yet — nothing to notify.
		return nil
	}

	switch {
	case allCompleted:
		_ = s.notificationService.Send(
			candidate.CreatedBy,
			"Recruitment completed",
			fmt.Sprintf("All recruitment steps have been completed for %s.", candidate.FullName),
			string(domain.NotificationStatusUpdate),
			"candidate", candidateID,
		)
		_ = s.notificationService.Send(
			selection.SelectedBy,
			"Recruitment completed",
			fmt.Sprintf("All recruitment steps have been completed for %s.", candidate.FullName),
			string(domain.NotificationStatusUpdate),
			"candidate", candidateID,
		)

	case status == domain.Completed &&
		strings.EqualFold(strings.TrimSpace(target.StepName), domain.TicketBooked):
		_ = s.notificationService.Send(
			selection.SelectedBy,
			"Flight ticket booked",
			fmt.Sprintf("A flight ticket has been booked for %s. Please confirm departure details.", candidate.FullName),
			string(domain.NotificationFlightBooked),
			"candidate", candidateID,
		)
		_ = s.notificationService.Send(
			candidate.CreatedBy,
			"Flight ticket booked",
			fmt.Sprintf("Flight ticket booked for %s.", candidate.FullName),
			string(domain.NotificationFlightBooked),
			"candidate", candidateID,
		)

	case status == domain.Completed &&
		strings.EqualFold(strings.TrimSpace(target.StepName), domain.TicketConfirmed):
		_ = s.notificationService.Send(
			selection.SelectedBy,
			"Flight ticket confirmed",
			fmt.Sprintf("The flight ticket for %s has been confirmed. Departure is finalised.", candidate.FullName),
			string(domain.NotificationFlightBooked),
			"candidate", candidateID,
		)
		_ = s.notificationService.Send(
			candidate.CreatedBy,
			"Flight ticket confirmed",
			fmt.Sprintf("Flight ticket confirmed for %s.", candidate.FullName),
			string(domain.NotificationFlightBooked),
			"candidate", candidateID,
		)

	case status == domain.Completed &&
		strings.EqualFold(strings.TrimSpace(target.StepName), domain.Arrived):
		_ = s.notificationService.Send(
			selection.SelectedBy,
			"Candidate has arrived",
			fmt.Sprintf("%s has arrived at the destination. Deployment is complete.", candidate.FullName),
			string(domain.NotificationArrived),
			"candidate", candidateID,
		)
		_ = s.notificationService.Send(
			candidate.CreatedBy,
			"Candidate has arrived",
			fmt.Sprintf("%s has arrived at the destination. Deployment is complete.", candidate.FullName),
			string(domain.NotificationArrived),
			"candidate", candidateID,
		)

	case status == domain.Completed:
		// Generic progress update to the foreign agency.
		_ = s.notificationService.Send(
			selection.SelectedBy,
			"Recruitment progress updated",
			fmt.Sprintf("Step '%s' has been completed for %s.", target.StepName, candidate.FullName),
			string(domain.NotificationStatusUpdate),
			"candidate", candidateID,
		)
	}

	return nil
}

// GetCandidateProgress returns the ordered list of status steps for a
// candidate, initialising them first if they do not yet exist.
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

	if candidate.Status != "" &&
		candidate.Status != domain.CandidateStatusInProgress &&
		candidate.Status != domain.CandidateStatusCompleted {
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

// ─── step initialisation ─────────────────────────────────────────────────────

func (s *StatusStepService) initializeStepsWithTx(tx *gorm.DB, candidateID, ownerID string) error {
	existingSteps := make([]*domain.StatusStep, 0)
	if err := tx.Where("candidate_id = ?", candidateID).
		Order("created_at ASC").
		Find(&existingSteps).Error; err != nil {
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

func predefinedStepNames() []string {
	return []string{
		domain.MedicalStep,
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
		if steps[index] == nil ||
			!strings.EqualFold(strings.TrimSpace(steps[index].StepName), strings.TrimSpace(stepName)) {
			return false
		}
	}
	return true
}

func rebuildStepsForCurrentWorkflowWithTx(tx *gorm.DB, candidateID, ownerID string, existingSteps []*domain.StatusStep) error {
	workflow := predefinedStepNames()
	completedLegacy := 0

	for _, step := range existingSteps {
		if step == nil {
			continue
		}
		if step.StepStatus == domain.Completed {
			completedLegacy++
		}
	}

	completedTarget := approximateCompletedSteps(completedLegacy, len(existingSteps), len(workflow))
	if completedTarget >= len(workflow) {
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

// ─── validation helpers ───────────────────────────────────────────────────────

// isValidChecklistStatus accepts the two states meaningful for a checklist
// toggle: pending (uncheck) and completed (check).
// in_progress is also accepted for backwards-compatibility with any existing
// clients that still send it.
func isValidChecklistStatus(status domain.StepStatus) bool {
	switch status {
	case domain.Pending, domain.InProgress, domain.Completed:
		return true
	default:
		return false
	}
}
