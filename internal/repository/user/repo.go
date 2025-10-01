package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aliskhannn/calendar-service/internal/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// Repository manages interactions with the users table in the PostgreSQL database.
// It provides methods for creating and retrieving user records.
type Repository struct {
	db *pgxpool.Pool // Database connection pool
}

// New creates a new Repository instance with the provided database connection pool.
//
// Parameters:
//   - db: The PostgreSQL connection pool for database operations.
//
// Returns:
//   - A pointer to the initialized Repository.
func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// CreateUser inserts a new user into the users table and returns their ID.
// It stores the user's name, email, and password hash.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - user: The user data to be inserted.
//
// Returns:
//   - The UUID of the created user.
//   - An error if the insertion fails.
func (r *Repository) CreateUser(ctx context.Context, user model.User) (uuid.UUID, error) {
	query := `
		INSERT INTO users (
		    name, email, password_hash
		) VALUES ($1, $2, $3)
		RETURNING id
   `

	err := r.db.QueryRow(
		ctx, query, user.Name, user.Email, user.Password,
	).Scan(&user.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user.ID, nil
}

// GetUserByID retrieves a user from the users table by their ID.
// It returns the user's details, including ID, email, name, password hash, and timestamps.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - id: The UUID of the user to retrieve.
//
// Returns:
//   - A pointer to the retrieved user.
//   - An error if the query fails or if the user is not found.
func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
   `

	var user model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user from the users table by their email address.
// It returns the user's details, including ID, email, name, password hash, and timestamps.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - email: The email address of the user to retrieve.
//
// Returns:
//   - A pointer to the retrieved user.
//   - An error if the query fails or if the user is not found.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
   `

	var user model.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}
