package jobs

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"maid-recruitment-tracking/internal/domain"
)

// ─── interfaces ───────────────────────────────────────────────────────────────

// selectionWarningRepo is the subset of SelectionRepository used by the warning job.
type selectionWarningRepo interface {
	GetPendingExpiringSoon(windowHours int, flagBit int) ([]*domain.Selection, error)
	UpdateWarningSentFlags(id string, flags int) error
	GetByID(id string) (*domain.Selection, error)
}

// passportWarningRepo is the subset of PassportDataRepository used by the warning job.
type passportWarningRepo interface {
	GetExpiringPassports(daysAhead int, flagBit int) ([]*domain.PassportData, error)
	UpdateWarningSentFlags(id string, flags int) error
}

// candidateNameRepo fetches a candidate's name for notification messages.
type candidateNameRepo interface {
	GetByID(id string) (*domain.Candidate, error)
}

// warningNotifier is the subset of NotificationService used by the warning job.
type warningNotifier interface {
	Send(userID, title, message, notificationType, relatedEntityType, relatedEntityID string) error
}

// ─── job struct ───────────────────────────────────────────────────────────────

// ExpiryWarningJob fires pre-expiry notifications for:
//   - Selections expiring within 24 h, 6 h, and 1 h (once per level)
//   - Passport expiry within 6 months, 3 months, and 1 month (once per level)
type ExpiryWarningJob struct {
	selectionRepo selectionWarningRepo
	passportRepo  passportWarningRepo
	candidateRepo candidateNameRepo
	notifier      warningNotifier
}

// NewExpiryWarningJob constructs an ExpiryWarningJob.
// All four dependencies are required; an error is returned if any is nil.
func NewExpiryWarningJob(
	selectionRepo selectionWarningRepo,
	passportRepo passportWarningRepo,
	candidateRepo candidateNameRepo,
	notifier warningNotifier,
) (*ExpiryWarningJob, error) {
	if selectionRepo == nil {
		return nil, fmt.Errorf("expiry warning job: selectionRepo is nil")
	}
	if passportRepo == nil {
		return nil, fmt.Errorf("expiry warning job: passportRepo is nil")
	}
	if candidateRepo == nil {
		return nil, fmt.Errorf("expiry warning job: candidateRepo is nil")
	}
	if notifier == nil {
		return nil, fmt.Errorf("expiry warning job: notifier is nil")
	}
	return &ExpiryWarningJob{
		selectionRepo: selectionRepo,
		passportRepo:  passportRepo,
		candidateRepo: candidateRepo,
		notifier:      notifier,
	}, nil
}

// Run executes all warning checks. It is called by the cron scheduler and
// also safe to call manually in tests.
func (j *ExpiryWarningJob) Run() {
	j.processSelectionWarnings()
	j.processPassportWarnings()
}

// ─── selection warnings ───────────────────────────────────────────────────────

// selectionWarningLevel describes one pre-expiry warning tier.
type selectionWarningLevel struct {
	windowHours int    // look-ahead window: warn if expires within this many hours
	flagBit     int    // domain.WarningSent* constant
	label       string // human-readable label used in notification text
}

var selectionWarningLevels = []selectionWarningLevel{
	{windowHours: 25, flagBit: domain.WarningSent24h, label: "24 hours"},
	{windowHours: 7, flagBit: domain.WarningSent6h, label: "6 hours"},
	{windowHours: 2, flagBit: domain.WarningSent1h, label: "1 hour"},
}

func (j *ExpiryWarningJob) processSelectionWarnings() {
	for _, level := range selectionWarningLevels {
		selections, err := j.selectionRepo.GetPendingExpiringSoon(level.windowHours, level.flagBit)
		if err != nil {
			log.Printf("expiry warning job: fetch selections (window=%dh): %v", level.windowHours, err)
			continue
		}

		for _, sel := range selections {
			if sel == nil {
				continue
			}
			j.sendSelectionWarning(sel, level)
		}
	}
}

func (j *ExpiryWarningJob) sendSelectionWarning(sel *domain.Selection, level selectionWarningLevel) {
	// Resolve candidate name for a richer notification message.
	candidateName := "the candidate"
	candidate, err := j.candidateRepo.GetByID(sel.CandidateID)
	if err == nil && candidate != nil && strings.TrimSpace(candidate.FullName) != "" {
		candidateName = candidate.FullName
	}

	// Calculate remaining time for the message.
	remaining := time.Until(sel.ExpiresAt)
	remainingStr := formatDuration(remaining)

	title := fmt.Sprintf("Selection expiring in %s", level.label)
	message := fmt.Sprintf(
		"The selection for %s expires in %s (%s). Review and approve before it expires.",
		candidateName,
		level.label,
		remainingStr,
	)

	// Notify the Ethiopian agency (candidate owner).
	if strings.TrimSpace(candidate.CreatedBy) != "" {
		_ = j.notifier.Send(
			candidate.CreatedBy,
			title,
			message,
			string(domain.NotificationExpiryWarning),
			"selection",
			sel.ID,
		)
	}

	// Notify the foreign agency (selector).
	if strings.TrimSpace(sel.SelectedBy) != "" {
		_ = j.notifier.Send(
			sel.SelectedBy,
			title,
			message,
			string(domain.NotificationExpiryWarning),
			"selection",
			sel.ID,
		)
	}

	// Mark the flag so this level is not sent again.
	newFlags := sel.WarningSentFlags | level.flagBit
	if updateErr := j.selectionRepo.UpdateWarningSentFlags(sel.ID, newFlags); updateErr != nil {
		log.Printf("expiry warning job: update selection flags (id=%s): %v", sel.ID, updateErr)
	}
}

// ─── passport warnings ────────────────────────────────────────────────────────

// passportWarningLevel describes one passport-expiry warning tier.
type passportWarningLevel struct {
	daysAhead int    // warn when expiry_date <= now + daysAhead
	flagBit   int    // domain.PassportWarning* constant
	label     string // human-readable label used in notification text
}

var passportWarningLevels = []passportWarningLevel{
	{daysAhead: 180, flagBit: domain.PassportWarning6Months, label: "6 months"},
	{daysAhead: 90, flagBit: domain.PassportWarning3Months, label: "3 months"},
	{daysAhead: 30, flagBit: domain.PassportWarning1Month, label: "1 month"},
}

func (j *ExpiryWarningJob) processPassportWarnings() {
	for _, level := range passportWarningLevels {
		passports, err := j.passportRepo.GetExpiringPassports(level.daysAhead, level.flagBit)
		if err != nil {
			log.Printf("expiry warning job: fetch expiring passports (days=%d): %v", level.daysAhead, err)
			continue
		}

		for _, pd := range passports {
			if pd == nil {
				continue
			}
			j.sendPassportWarning(pd, level)
		}
	}
}

func (j *ExpiryWarningJob) sendPassportWarning(pd *domain.PassportData, level passportWarningLevel) {
	// Resolve candidate name and owner.
	candidateName := "a candidate"
	ownerID := ""

	candidate, err := j.candidateRepo.GetByID(pd.CandidateID)
	if err == nil && candidate != nil {
		if strings.TrimSpace(candidate.FullName) != "" {
			candidateName = candidate.FullName
		}
		ownerID = strings.TrimSpace(candidate.CreatedBy)
	}

	if ownerID == "" {
		log.Printf("expiry warning job: cannot resolve owner for passport data id=%s", pd.ID)
		return
	}

	daysLeft := pd.DaysUntilExpiry()
	var daysStr string
	if daysLeft <= 0 {
		daysStr = "already expired"
	} else {
		daysStr = fmt.Sprintf("%d days", daysLeft)
	}

	title := fmt.Sprintf("Passport expiring in %s", level.label)
	message := fmt.Sprintf(
		"The passport for %s expires in approximately %s (%s remaining). "+
			"Renewal is required before deployment can proceed.",
		candidateName,
		level.label,
		daysStr,
	)

	_ = j.notifier.Send(
		ownerID,
		title,
		message,
		string(domain.NotificationPassportExpiry),
		"candidate",
		pd.CandidateID,
	)

	// Mark the flag so this level is not resent.
	newFlags := pd.PassportWarningSentFlags | level.flagBit
	if updateErr := j.passportRepo.UpdateWarningSentFlags(pd.ID, newFlags); updateErr != nil {
		log.Printf("expiry warning job: update passport warning flags (id=%s): %v", pd.ID, updateErr)
	}
}

// ─── scheduler ────────────────────────────────────────────────────────────────

// StartExpiryWarningScheduler registers the ExpiryWarningJob on a cron
// scheduler that runs every 30 minutes and starts it.
// The caller must call scheduler.Stop() on shutdown.
func StartExpiryWarningScheduler(
	selectionRepo selectionWarningRepo,
	passportRepo passportWarningRepo,
	candidateRepo candidateNameRepo,
	notifier warningNotifier,
) (*cron.Cron, error) {
	job, err := NewExpiryWarningJob(selectionRepo, passportRepo, candidateRepo, notifier)
	if err != nil {
		return nil, err
	}

	scheduler := cron.New(cron.WithSeconds())
	// "0 */30 * * * *" = at second 0 of every 30th minute
	if _, err := scheduler.AddJob("0 */30 * * * *", job); err != nil {
		return nil, fmt.Errorf("register expiry warning job: %w", err)
	}
	scheduler.Start()

	return scheduler, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// formatDuration converts a duration into a short human-readable string
// like "5h 22m" or "47m".
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "less than a minute"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 && m > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if h > 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dm", m)
}
