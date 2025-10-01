package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/aliskhannn/calendar-service/internal/api/handlers/auth"
	"github.com/aliskhannn/calendar-service/internal/api/handlers/event"
	"github.com/aliskhannn/calendar-service/internal/config"
	"github.com/aliskhannn/calendar-service/internal/middlewares"
)

func New(authHandler *auth.Handler, eventHandler *event.Handler, config *config.Config, logCh chan<- middlewares.LogEntry) http.Handler {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(middlewares.Logger(logCh))

	authMiddleware := middlewares.Auth(config.JWT)

	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			r.Route("/events", func(r chi.Router) {
				r.Post("/", eventHandler.Create)
				r.Put("/{id}", eventHandler.Update)
				r.Delete("/{id}", eventHandler.Delete)

				r.Get("/day", eventHandler.GetDay)
				r.Get("/week", eventHandler.GetWeek)
				r.Get("/month", eventHandler.GetMonth)
			})
		})
	})

	return r
}
