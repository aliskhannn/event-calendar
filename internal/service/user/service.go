package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aliskhannn/calendar-service/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/aliskhannn/calendar-service/internal/model"
	userrepo "github.com/aliskhannn/calendar-service/internal/repository/user"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

//go:generate mockgen -source=service.go -destination=../../mocks/service/user/mock_user.go -package=mocks
type userRepository interface {
	CreateUser(ctx context.Context, user model.User) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

type Service struct {
	userRepo userRepository
	config   *config.Config
}

func New(userRepo userRepository, config *config.Config) *Service {
	return &Service{
		userRepo: userRepo,
		config:   config,
	}
}

func (s *Service) Create(ctx context.Context, email, name, password string) (uuid.UUID, error) {
	// Check if user already exists.
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return uuid.Nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, userrepo.ErrUserNotFound) {
		return uuid.Nil, fmt.Errorf("get user by email: %w", err)
	}

	// Hash password.
	hash, err := hashPassword(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		Email:    email,
		Name:     name,
		Password: hash,
	}

	id, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create user: %w", err)
	}

	return id, nil
}

func (s *Service) GetByEmail(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userrepo.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}

		return "", fmt.Errorf("get user by email: %w", err)
	}

	// Verify password.
	if err := verifyPassword(password, user.Password); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token.
	token, err := generateToken(user, s.config.JWT)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

// hashPassword generates a bcrypt hash for the given password.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hash), err
}

// verifyPassword verifies if the given password matches the stored hash.
func verifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// generateToken creates a token for user.
func generateToken(user *model.User, jwtCfg config.JWT) (string, error) {
	expTime := time.Now().Add(jwtCfg.TTL)

	// Create the JWT claims.
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"name":    user.Name,
		"email":   user.Email,
		"exp":     expTime.Unix(),    // expiration time
		"iat":     time.Now().Unix(), // issued at time
	}

	// Create the token with claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret key and return.
	return token.SignedString([]byte(jwtCfg.Secret))
}
