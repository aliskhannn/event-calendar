package middlewares

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

func Logger(logger *zap.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("request",
				zap.String("method", r.Method),
				zap.String("url", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
