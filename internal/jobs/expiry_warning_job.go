package jobs

import (
	"fmt"
	"time"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
)

const (
	selectionWarning24Hours = 1 << 0
	selectionWarning6Hours  = 1 << 1
	selectionWarning1Hour   = 1 << 2

	passportWarning6Months = 1 << 0
	passportWarning3Months = 1 << 1
	passportWarning1Month  = 1 << 2

	medicalWarning30Days = 1 << 0
	medicalWarning14Days = 1 << 1
	medicalWarning7Days  = 1 << 2
)

type ExpiryWarningJob struct {
	selectionRepository *repository.GormSelectionRepository
	candidateRepository domain.CandidateRepository
	passportRepository  *repository.GormPassportDataRepository
	medicalRepository   *repository.GormMedicalDataRepository
	notificationService *service.NotificationService
}

func NewExpiryWarningJob(
	selectionRepository *repository.GormSelectionRepository,
	candidateRepository domain.CandidateRepository,
	passportRepository *repository.GormPassportDataRepository,
	medicalRepository *repository.GormMedicalDataRepository,
	notificationService *service.NotificationService,
) (*ExpiryWarningJob, error) {
	if selectionRepository == nil {
		return nil, fmt.Errorf("selection repository is nil")
	}
	if candidateRepository == nil {
		return nil, fmt.Errorf("candidate repository is nil")
	}
	if passportRepository == nil {
		return nil, fmt.Errorf("passport repository is nil")
	}
	if medicalRepository == nil {
		return nil, fmt.Errorf("medical repository is nil")
	}
	if notificationService == nil {
		return nil, fmt.Errorf("notification service is nil")
	}

	return &ExpiryWarningJob{
		selectionRepository: selectionRepository,
		candidateRepository: candidateRepository,
		passportRepository:  passportRepository,
		medicalRepository:   medicalRepository,
		notificationService: notificationService,
	}, nil
}

func (j *ExpiryWarningJob) Run() {
	_ = j.processSelectionWarnings()
	_ = j.processPassportWarnings()
	_ = j.processMedicalWarnings()
}

func (j *ExpiryWarningJob) processSelectionWarnings() error {
	now := time.Now().UTC()
	deadline := now.Add(24 * time.Hour)

	selections := make([]*domain.Selection, 0)
	if err := j.selectionRepository.DB().
		Where("status = ? AND expires_at > ? AND expires_at <= ?", domain.SelectionPending, now, deadline).
		Find(&selections).Error; err != nil {
		return fmt.Errorf("load expiring selections: %w", err)
	}

	for _, selection := range selections {
		if selection == nil {
			continue
		}

		remaining := time.Until(selection.ExpiresAt.UTC())
		flags := selection.WarningSentFlags

		thresholds := []struct {
			bit   int
			limit time.Duration
			label string
		}{
			{bit: selectionWarning24Hours, limit: 24 * time.Hour, label: "24 hours"},
			{bit: selectionWarning6Hours, limit: 6 * time.Hour, label: "6 hours"},
			{bit: selectionWarning1Hour, limit: time.Hour, label: "1 hour"},
		}

		updatedFlags := flags
		for _, threshold := range thresholds {
			if remaining > threshold.limit || updatedFlags&threshold.bit != 0 {
				continue
			}
			if err := j.notificationService.NotifyExpiryWarning(selection.ID, threshold.label); err != nil {
				return err
			}
			updatedFlags |= threshold.bit
		}

		if updatedFlags != flags {
			if err := j.selectionRepository.DB().
				Model(&domain.Selection{}).
				Where("id = ?", selection.ID).
				Update("warning_sent_flags", updatedFlags).Error; err != nil {
				return fmt.Errorf("persist selection warning flags: %w", err)
			}
		}
	}

	return nil
}

func (j *ExpiryWarningJob) processPassportWarnings() error {
	passports, err := j.passportRepository.GetExpiringPassports(180)
	if err != nil {
		return err
	}

	for _, passport := range passports {
		if passport == nil {
			continue
		}

		candidate, err := j.candidateRepository.GetByID(passport.CandidateID)
		if err != nil {
			return err
		}

		remaining := time.Until(passport.ExpiryDate.UTC())
		flags := passport.PassportWarningSentFlags
		thresholds := []struct {
			bit   int
			limit time.Duration
			label string
		}{
			{bit: passportWarning6Months, limit: 180 * 24 * time.Hour, label: "6 months"},
			{bit: passportWarning3Months, limit: 90 * 24 * time.Hour, label: "3 months"},
			{bit: passportWarning1Month, limit: 30 * 24 * time.Hour, label: "1 month"},
		}

		updatedFlags := flags
		for _, threshold := range thresholds {
			if remaining > threshold.limit || updatedFlags&threshold.bit != 0 {
				continue
			}
			if err := j.notificationService.NotifyPassportExpiry(passport.CandidateID, candidate.CreatedBy, threshold.label); err != nil {
				return err
			}
			updatedFlags |= threshold.bit
		}

		if updatedFlags != flags {
			if err := j.passportRepository.DB().
				Model(&domain.PassportData{}).
				Where("id = ?", passport.ID).
				Update("passport_warning_sent_flags", updatedFlags).Error; err != nil {
				return fmt.Errorf("persist passport warning flags: %w", err)
			}
		}
	}

	return nil
}

func (j *ExpiryWarningJob) processMedicalWarnings() error {
	records, err := j.medicalRepository.GetExpiringMedical(30)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record == nil {
			continue
		}

		candidate, err := j.candidateRepository.GetByID(record.CandidateID)
		if err != nil {
			return err
		}

		remaining := time.Until(record.ExpiryDate.UTC())
		flags := record.WarningSentFlags
		thresholds := []struct {
			bit   int
			limit time.Duration
			label string
		}{
			{bit: medicalWarning30Days, limit: 30 * 24 * time.Hour, label: "30 days"},
			{bit: medicalWarning14Days, limit: 14 * 24 * time.Hour, label: "14 days"},
			{bit: medicalWarning7Days, limit: 7 * 24 * time.Hour, label: "7 days"},
		}

		updatedFlags := flags
		for _, threshold := range thresholds {
			if remaining > threshold.limit || updatedFlags&threshold.bit != 0 {
				continue
			}
			if err := j.notificationService.NotifyMedicalExpiry(record.CandidateID, candidate.CreatedBy, threshold.label); err != nil {
				return err
			}
			updatedFlags |= threshold.bit
		}

		if updatedFlags != flags {
			if err := j.medicalRepository.DB().
				Model(&domain.MedicalData{}).
				Where("id = ?", record.ID).
				Update("warning_sent_flags", updatedFlags).Error; err != nil {
				return fmt.Errorf("persist medical warning flags: %w", err)
			}
		}
	}

	return nil
}
