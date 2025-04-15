package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

func WithAuth(log *zap.Logger, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string

			cookie, err := r.Cookie("auth")
			if err == nil {
				userID, err = getUserID(cookie.Value, secret)
			}
			if err != nil {
				userID = uuid.NewString()
				token, err := generateToken(secret, userID)
				if err != nil {
					log.Error("failed to generate JWT", zap.Error(err))
					http.Error(w, "authentication error", http.StatusInternalServerError)
					return
				}
				newCookie := &http.Cookie{
					Name:     "auth",
					Value:    token,
					Path:     "/",
					HttpOnly: true,
				}
				http.SetCookie(w, newCookie)
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func generateToken(secret string, userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// реализовать вместе с механизмом обновления токенов ?
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		UserID: userID,
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getUserID(tokenString string, secret string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	return claims.UserID, nil
}
