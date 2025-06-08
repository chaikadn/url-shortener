package generator

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const keyLen = 8

type Generator interface {
	Generate(url string) (key string, err error)
	// IsValid(key string) bool
}

type Crypto struct{}

func NewCrypto() *Crypto {
	return &Crypto{}
}

func (c *Crypto) Generate(url string) (string, error) {
	result := make([]byte, keyLen)

	for i := range keyLen {
		indexBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random key: %w", err)
		}
		index := indexBig.Int64()
		result[i] = alphabet[index]
	}
	return string(result), nil
}
