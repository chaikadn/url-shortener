package random

import (
	"crypto/rand"
	"fmt"
)

// использовать base62 генерацию
func RandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	for i := range length {
		randomBytes[i] = charset[randomBytes[i]%byte(len(charset))]
	}
	return string(randomBytes), nil
}
