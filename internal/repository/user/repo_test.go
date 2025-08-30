//go:build integration
// +build integration

package user

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/aliskhannn/calendar-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	testRepo   *Repository
	testUserID uuid.UUID
)

func TestMain(m *testing.M) {
	_ = godotenv.Load(".env.test")

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		panic("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}

	testRepo = New(db)
	testUserID = uuid.New()

	if _, err := db.Exec(context.Background(), "DELETE FROM users"); err != nil {
		panic(err)
	}

	code := m.Run()
	db.Close()
	os.Exit(code)
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	user := model.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "secret123",
	}

	id, err := testRepo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id == uuid.Nil {
		t.Fatal("expected valid UUID, got Nil")
	}
}

func TestGetUserByEmail(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"

	u, err := testRepo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if u.Email != email {
		t.Fatalf("expected email %s, got %s", email, u.Email)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	_, err := testRepo.GetUserByEmail(ctx, "notfound@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
