package main

import (
	"context"
	"errors"
	authhandler "github.com/aliskhannn/calendar-service/internal/api/handlers/auth"
	eventhandler "github.com/aliskhannn/calendar-service/internal/api/handlers/event"
	"github.com/aliskhannn/calendar-service/internal/api/router"
	"github.com/aliskhannn/calendar-service/internal/api/server"
	"github.com/aliskhannn/calendar-service/internal/config"
	"github.com/aliskhannn/calendar-service/internal/logger"
	eventrepo "github.com/aliskhannn/calendar-service/internal/repository/event"
	userrepo "github.com/aliskhannn/calendar-service/internal/repository/user"
	eventsvc "github.com/aliskhannn/calendar-service/internal/service/event"
	usersvc "github.com/aliskhannn/calendar-service/internal/service/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Must()
	log := logger.CreateLogger()
	val := validator.New()

	dbpool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatal("error creating connection pool", zap.Error(err))
	}

	eventRepo := eventrepo.New(dbpool)
	userRepo := userrepo.New(dbpool)

	eventSvc := eventsvc.New(eventRepo)
	eventHandler := eventhandler.New(eventSvc, log, val)

	userSvc := usersvc.New(userRepo, cfg)
	authHandler := authhandler.New(userSvc, log, val)

	r := router.New(authHandler, eventHandler, cfg, log)
	s := server.New(cfg.Server.HTTPPort, r)

	go func() {
		log.Info("starting HTTP server", zap.String("port", cfg.Server.HTTPPort))
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

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
	dbpool.Close()
}
