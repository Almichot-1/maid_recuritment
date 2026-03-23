package jobs

import (
	"log"

	"github.com/robfig/cron/v3"

	"maid-recruitment-tracking/internal/service"
)

type ExpiryJob struct {
	selectionService *service.SelectionService
}

func NewExpiryJob(selectionService *service.SelectionService) (*ExpiryJob, error) {
	if selectionService == nil {
		return nil, ErrNilSelectionService
	}
	return &ExpiryJob{selectionService: selectionService}, nil
}

func (j *ExpiryJob) Run() {
	if err := j.selectionService.ProcessExpiredSelections(); err != nil {
		log.Printf("expiry job failed: %v", err)
	}
}

func StartExpiryScheduler(selectionService *service.SelectionService) (*cron.Cron, error) {
	job, err := NewExpiryJob(selectionService)
	if err != nil {
		return nil, err
	}

	scheduler := cron.New(cron.WithSeconds())
	if _, err := scheduler.AddJob("0 */5 * * * *", job); err != nil {
		return nil, err
	}
	scheduler.Start()

	return scheduler, nil
}
