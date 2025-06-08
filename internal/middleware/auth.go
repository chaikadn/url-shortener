package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/chaikadn/url-shortener/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken            = errors.New("invalid token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidAuthHeader       = errors.New("invalid authorization header format")
)

type contextKey string

var (
	UserIDkey contextKey = "user_id"
	// RoleKey   contextKey = "role"
)

// TODO: выводить ошибки в JSON
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")
			if authorization == "" {
				http.Error(w, "authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString, err := extractToken(authorization)
			if err != nil {
				http.Error(w, "invalid authorization format", http.StatusUnauthorized)
				return
			}

			claims, err := validateJWT(tokenString, secret)
			if err != nil {

				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDkey, claims.UserID)
			// ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(header string) (string, error) {
	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidAuthHeader
	}
	return parts[1], nil
}

// добавить TTL токена
func validateJWT(tokenString string, JWTSecret string) (*model.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&model.JWTClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, t.Header["alg"])
			}
			return []byte(JWTSecret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*model.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
