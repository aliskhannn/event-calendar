package reminder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aliskhannn/calendar-service/internal/model"
)

// userService defines an interface for fetching user details.
type userService interface {
	// GetByID retrieves a user by their unique ID.
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// notifier defines an interface for sending notifications through a channel.
type notifier interface {
	// Send sends a notification message to the specified recipient.
	Send(to string, msg string) error
}

// Worker is responsible for processing reminders from the channel
// and sending notifications at the scheduled time.
type Worker struct {
	ch          <-chan model.Reminder // channel with reminders
	userService userService           // service to fetch user info
	notifier    notifier              // interface to send notifications
	logger      *zap.Logger           // structured logger
	wg          sync.WaitGroup        // wait group for active reminder goroutines
}

// NewWorker creates a new reminder worker.
func NewWorker(
	ch <-chan model.Reminder,
	userService userService,
	notifier notifier,
	l *zap.Logger,
) *Worker {
	return &Worker{
		ch:          ch,
		userService: userService,
		notifier:    notifier,
		logger:      l,
	}
}

// Start begins processing reminders in the background.
// It listens to the reminder channel and launches a goroutine for each reminder.
func (w *Worker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case r, ok := <-w.ch:
				if !ok {
					// Channel closed, wait for all active reminders to finish.
					w.wg.Wait()
					return
				}

				w.wg.Add(1)
				go w.handleReminder(ctx, r) // process reminder concurrently
			case <-ctx.Done():
				// Context cancelled, wait for all active reminders to finish.
				w.wg.Wait()
				return
			}
		}
	}()
}

// handleReminder waits until the scheduled reminder time and sends the notification.
func (w *Worker) handleReminder(ctx context.Context, r model.Reminder) {
	defer w.wg.Done()

	duration := time.Until(r.RemindAt)
	if duration > 0 {
		select {
		case <-time.After(duration):
			// Time to send the reminder.
		case <-ctx.Done():
			// Context cancelled before reminder time.
			return
		}
	}

	user, err := w.userService.GetByID(ctx, r.UserID)
	if err != nil {
		w.logger.Warn("failed to fetch user", zap.Error(err))
		return
	}

	reminderMsg := fmt.Sprintf("🔔 Reminder: your event \"%s\" is coming up!", r.Message)
	if err := w.notifier.Send(user.Email, reminderMsg); err != nil {
		w.logger.Warn("failed to send reminder message", zap.Error(err))
	}
}

// Stop waits for all active reminder goroutines to finish.
// Useful for graceful shutdown.
func (w *Worker) Stop() {
	w.wg.Wait()
}
