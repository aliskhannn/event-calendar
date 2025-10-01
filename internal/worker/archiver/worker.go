package archiver

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type eventService interface {
	ArchiveOldEvents(ctx context.Context) error
}

type Worker struct {
	eventService eventService
	logger       *zap.Logger // structured logger
}

func NewWorker(eventService eventService, l *zap.Logger) *Worker {
	return &Worker{
		eventService: eventService,
		logger:       l,
	}
}

func (w *Worker) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := w.eventService.ArchiveOldEvents(ctx); err != nil {
					w.logger.Error("failed to archive old events", zap.Error(err))
				} else {
					w.logger.Info("successfully archived old events")
				}
			case <-ctx.Done():
				w.logger.Info("archiver worker stopped")
				return
			}
		}
	}()
}
