package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aliskhannn/delayed-notifier/pkg/email"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	authhandler "github.com/aliskhannn/calendar-service/internal/api/handlers/auth"
	eventhandler "github.com/aliskhannn/calendar-service/internal/api/handlers/event"
	"github.com/aliskhannn/calendar-service/internal/api/router"
	"github.com/aliskhannn/calendar-service/internal/api/server"
	"github.com/aliskhannn/calendar-service/internal/config"
	"github.com/aliskhannn/calendar-service/internal/logger"
	"github.com/aliskhannn/calendar-service/internal/middlewares"
	"github.com/aliskhannn/calendar-service/internal/model"
	eventrepo "github.com/aliskhannn/calendar-service/internal/repository/event"
	userrepo "github.com/aliskhannn/calendar-service/internal/repository/user"
	eventsvc "github.com/aliskhannn/calendar-service/internal/service/event"
	usersvc "github.com/aliskhannn/calendar-service/internal/service/user"
	"github.com/aliskhannn/calendar-service/internal/worker/archiver"
	"github.com/aliskhannn/calendar-service/internal/worker/reminder"
)

func main() {
	// Context for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load configuration.
	cfg := config.Must()

	// Initialize logger and validator.
	log := logger.CreateLogger()
	val := validator.New()

	// Connect to database.
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatal("error creating connection pool", zap.Error(err))
	}

	// Repositories.
	userRepo := userrepo.New(dbPool)
	eventRepo := eventrepo.New(dbPool)

	// Repositories.
	userSvc := usersvc.New(userRepo, cfg)
	eventSvc := eventsvc.New(eventRepo)

	// Reminder channel.
	reminderCh := make(chan model.Reminder, 100)

	// HTTP Handlers.
	authHandler := authhandler.New(userSvc, log, val)
	eventHandler := eventhandler.New(eventSvc, reminderCh, log, val)

	// Email client for reminders.
	smtpPort, err := strconv.Atoi(cfg.Email.SMTPPort)
	if err != nil {
		log.Fatal("error parsing SMTP port", zap.Error(err))
	}

	emailClient := email.NewClient(
		cfg.Email.SMTPHost,
		smtpPort,
		cfg.Email.Username,
		cfg.Email.Password,
		cfg.Email.From,
	)

	// Start reminder worker.
	reminderWorker := reminder.NewWorker(reminderCh, userSvc, emailClient, log)
	reminderWorker.Start(ctx)

	// Start archiver worker.
	archiverWorker := archiver.NewWorker(eventSvc, log)
	archiverWorker.Start(ctx, cfg.Archiver.Interval)

	// Async logging.
	logCh := make(chan middlewares.LogEntry, 100)
	middlewares.StartAsyncLogger(logCh, log)

	// Setup router and server.
	r := router.New(authHandler, eventHandler, cfg, logCh)
	s := server.New(cfg.Server.HTTPPort, r)

	go func() {
		log.Info("starting HTTP server", zap.String("port", cfg.Server.HTTPPort))
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	log.Info("shutdown signal received")

	// Graceful shutdown with timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("shutting down HTTP server...")
	if err = s.Shutdown(shutdownCtx); err != nil {
		log.Error("could not shutdown HTTP server", zap.Error(err))
	}

	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		log.Fatal("timeout exceeded, forcing shutdown")
	}

	log.Info("closing database pool...")
	dbPool.Close()
}
