package service

import (
	"fmt"
	"time"

	"github.com/chaikadn/url-shortener/internal/model"
	"github.com/chaikadn/url-shortener/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(username string, password string) (userID int, err error)
	Login(username string, password string, secret string) (token string, err error)
	GetUser(userID int) (*model.User, error)
}

type userService struct {
	db storage.UserStorage
}

func NewUserService(db storage.UserStorage) UserService {
	return &userService{db: db}
}

func (u *userService) Register(username string, password string) (int, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return -1, fmt.Errorf("failed to hash password: %w", err)
	}
	user := &model.User{
		Name:         username,
		Role:         "user",
		PasswordHash: passwordHash,
	}
	userID, err := u.db.SaveUser(user)
	if err != nil {
		return -1, fmt.Errorf("failed to create user: %w", err)
	}
	return userID, nil
}

func (u *userService) Login(username string, password string, secret string) (string, error) {
	user, err := u.db.GetUserByName(username)
	if err != nil {
		return "", fmt.Errorf("failed to get user by name: %w", err)
	}

	if err := verifyPassword(user.PasswordHash, password); err != nil {
		return "", fmt.Errorf("failed to verify password: %w", err)
	}

	token, err := generateJWT(user.ID, secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

func (u *userService) GetUser(userID int) (*model.User, error) {
	user, err := u.db.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

func hashPassword(password string) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to make hash: %w", err)
	}
	return bytes, nil
}

func verifyPassword(hashedPassword []byte, password string) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
}

func generateJWT(userID int, secret string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &model.JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "user_token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
