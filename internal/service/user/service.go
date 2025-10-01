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

// userRepository defines the interface for user-related database operations.
// It provides methods for creating and retrieving users by ID or email.
type userRepository interface {
	// CreateUser inserts a new user into the database and returns their ID.
	CreateUser(ctx context.Context, user model.User) (uuid.UUID, error)

	// GetUserByID retrieves a user by their ID.
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// GetUserByEmail retrieves a user by their email address.
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

// Service manages business logic for user-related operations.
// It handles user creation, retrieval, and authentication, including password hashing and JWT generation.
type Service struct {
	userRepo userRepository // Repository for user database operations
	config   *config.Config // Application configuration, including JWT settings
}

// New creates a new Service instance with the provided user repository and configuration.
//
// Parameters:
//   - userRepo: The repository for user database operations.
//   - config: The application configuration containing JWT settings.
//
// Returns:
//   - A pointer to the initialized Service.
func New(userRepo userRepository, config *config.Config) *Service {
	return &Service{
		userRepo: userRepo,
		config:   config,
	}
}

// Create registers a new user with the provided email, name, and password.
// It checks if the email is already in use, hashes the password, and creates the user in the database.
//
// Parameters:
//   - ctx: The context for the operation.
//   - email: The email address of the new user.
//   - name: The name of the new user.
//   - password: The plaintext password to be hashed and stored.
//
// Returns:
//   - The UUID of the created user.
//   - An error if the email is already taken, password hashing fails, or user creation fails.
func (s *Service) Create(ctx context.Context, email, name, password string) (uuid.UUID, error) {
	// Check if user already exists.
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return uuid.Nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, userrepo.ErrUserNotFound) {
		return uuid.Nil, fmt.Errorf("get user by email: %w", err)
	}

	// Hash the password.
	hash, err := hashPassword(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		Email:    email,
		Name:     name,
		Password: hash,
	}

	// Create the user in the database.
	id, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create user: %w", err)
	}

	return id, nil
}

// GetByID retrieves a user by their ID.
// It fetches the user from the database and handles the case where the user is not found.
//
// Parameters:
//   - ctx: The context for the operation.
//   - id: The UUID of the user to retrieve.
//
// Returns:
//   - A pointer to the retrieved user.
//   - An error if the user is not found or the retrieval fails.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, userrepo.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail authenticates a user by their email and password, returning a JWT token if successful.
// It verifies the password and generates a JWT token with user details upon successful authentication.
//
// Parameters:
//   - ctx: The context for the operation.
//   - email: The email address of the user.
//   - password: The plaintext password to verify.
//
// Returns:
//   - A JWT token string if authentication is successful.
//   - An error if the user is not found, the password is invalid, or token generation fails.
func (s *Service) GetByEmail(ctx context.Context, email, password string) (string, error) {
	// Retrieve user by email.
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userrepo.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("get user by email: %w", err)
	}

	// Verify the password.
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
// It uses the default bcrypt cost for hashing.
//
// Parameters:
//   - password: The plaintext password to hash.
//
// Returns:
//   - The bcrypt hash as a string.
//   - An error if hashing fails.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// verifyPassword verifies if the given password matches the stored hash.
// It compares the plaintext password with the bcrypt hash.
//
// Parameters:
//   - password: The plaintext password to verify.
//   - hash: The stored bcrypt hash to compare against.
//
// Returns:
//   - An error if the password does not match the hash.
func verifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// generateToken creates a JWT token for the given user.
// It includes the user's ID, name, email, issuance time, and expiration time in the token claims.
//
// Parameters:
//   - user: The user for whom the token is generated.
//   - jwtCfg: The JWT configuration containing the secret and TTL.
//
// Returns:
//   - The signed JWT token string.
//   - An error if token generation or signing fails.
func generateToken(user *model.User, jwtCfg config.JWT) (string, error) {
	expTime := time.Now().Add(jwtCfg.TTL)

	// Create JWT claims.
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"name":    user.Name,
		"email":   user.Email,
		"exp":     expTime.Unix(),    // expiration time
		"iat":     time.Now().Unix(), // issued at time
	}

	// Create and sign the token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtCfg.Secret))
}
