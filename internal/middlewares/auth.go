package middlewares

import (
	"context"
	"errors"
	"github.com/aliskhannn/calendar-service/internal/api/response"
	"github.com/aliskhannn/calendar-service/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"os"
	"strings"
)

var (
	ErrNoToken            = errors.New("missing token")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidTokenFormat = errors.New("invalid token format")
	ErrExpiredToken       = errors.New("token had expired")
)

func Auth(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := r.Header.Get("Authorization")
			if tokenStr == "" {
				response.Fail(w, http.StatusUnauthorized, ErrNoToken)
				return
			}

			// Bearer token
			parts := strings.Split(tokenStr, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Fail(w, http.StatusUnauthorized, ErrInvalidTokenFormat)
				return
			}

			userID, err := validateToken(parts[1])
			if err != nil {
				response.Fail(w, http.StatusUnauthorized, ErrInvalidToken)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateToken verifies a JWT token and returns the claims.
func validateToken(tokenStr string) (uuid.UUID, error) {
	// Parse the token.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, ErrExpiredToken
		}

		return uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}
