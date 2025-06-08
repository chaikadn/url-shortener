package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ADD TAGS
type User struct {
	ID           int
	Name         string
	Role         string
	PasswordHash []byte
}

type URLEntry struct {
	ID          int
	Key         string
	OriginalURL string
	CreatedAt   time.Time
}

type JWTClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}
