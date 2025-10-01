package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// LogEntry defines a single log record for async logging.
type LogEntry struct {
	Method   string
	URL      string
	Duration time.Duration
	Time     time.Time
}

// StartAsyncLogger starts a background goroutine that reads from logCh
// and writes to the provided zap.Logger.
func StartAsyncLogger(logCh <-chan LogEntry, logger *zap.Logger) {
	go func() {
		for entry := range logCh {
			logger.Info("request",
				zap.String("method", entry.Method),
				zap.String("url", entry.URL),
				zap.Duration("duration", entry.Duration),
				zap.Time("time", entry.Time),
			)
		}
	}()
}

// Logger returns a middleware that sends log entries to logCh asynchronously.
func Logger(logCh chan<- LogEntry) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)

			// Send log entry to the channel.
			select {
			case logCh <- LogEntry{
				Method:   r.Method,
				URL:      r.URL.String(),
				Duration: time.Since(start),
				Time:     start,
			}:
			default:
				// If the channel is full, we drop the log to avoid blocking.
			}
		})
	}
}
