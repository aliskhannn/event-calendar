package archiver

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// eventService defines an interface for archiving old events.
type eventService interface {
	// ArchiveOldEvents moves old events to an archive or marks them as archived.
	ArchiveOldEvents(ctx context.Context) error
}

// Worker is responsible for periodically archiving old events.
type Worker struct {
	eventService eventService // service that performs the archiving
	logger       *zap.Logger  // structured logger
}

// NewWorker creates a new archiver worker.
func NewWorker(eventService eventService, l *zap.Logger) *Worker {
	return &Worker{
		eventService: eventService,
		logger:       l,
	}
}

// Start begins the archiving process.
// It runs a background goroutine that triggers ArchiveOldEvents
// at the specified interval. The goroutine stops gracefully when ctx is canceled.
func (w *Worker) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop() // stop the ticker when the goroutine exits

		for {
			select {
			case <-ticker.C:
				// Time to archive old events.
				if err := w.eventService.ArchiveOldEvents(ctx); err != nil {
					w.logger.Error("failed to archive old events", zap.Error(err))
				} else {
					w.logger.Info("successfully archived old events")
				}
			case <-ctx.Done():
				// Context cancelled, stop the worker gracefully.
				w.logger.Info("archiver worker stopped")
				return
			}
		}
	}()
}
