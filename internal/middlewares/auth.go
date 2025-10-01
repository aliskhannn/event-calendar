package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/aliskhannn/calendar-service/internal/api/response"
	"github.com/aliskhannn/calendar-service/internal/config"
)

var (
	ErrNoToken            = errors.New("missing token")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidTokenFormat = errors.New("invalid token format")
	ErrExpiredToken       = errors.New("token had expired")
)

// contextKey is a custom type to avoid collisions when storing values in context.
type contextKey string

// UserIDKey is the key used to store and retrieve the authenticated user's ID from the request context.
const UserIDKey contextKey = "user_id"

// Auth creates an HTTP middleware that enforces JWT authentication.
// It extracts and validates a JWT token from the Authorization header, verifies it using the provided secret,
// and stores the authenticated user ID in the request context if valid.
// If the token is missing, invalid, or expired, it returns an unauthorized response.
//
// Parameters:
//   - jwtCfg: The JWT configuration containing the secret key for token validation.
//
// Returns:
//   - An HTTP middleware handler that wraps the next handler in the chain.
func Auth(jwtCfg config.JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Authorization header.
			tokenStr := r.Header.Get("Authorization")
			if tokenStr == "" {
				response.Fail(w, http.StatusUnauthorized, ErrNoToken)
				return
			}

			// Validate Bearer token format.
			parts := strings.Split(tokenStr, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Fail(w, http.StatusUnauthorized, ErrInvalidTokenFormat)
				return
			}

			// Validate the JWT token and extract user ID.
			userID, err := validateToken(parts[1], jwtCfg.Secret)
			if err != nil {
				response.Fail(w, http.StatusUnauthorized, ErrInvalidToken)
				return
			}

			// Add user ID to request context and proceed to next handler.
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateToken verifies a JWT token and extracts the user ID from its claims.
// It checks the token's signing method, validity, and expiration, and parses the user ID from the claims.
//
// Parameters:
//   - tokenStr: The JWT token string to validate.
//   - secret: The secret key used to verify the token's signature.
//
// Returns:
//   - The user ID (UUID) extracted from the token claims.
//   - An error if the token is invalid, expired, or contains an invalid user ID.
func validateToken(tokenStr string, secret string) (uuid.UUID, error) {
	// Parse the token with the provided secret.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		// Handle expired token specifically.
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, ErrExpiredToken
		}
		return uuid.Nil, err
	}

	// Validate token and extract claims.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	// Extract and validate user ID from claims.
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}

	// Parse user ID into UUID.
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}
