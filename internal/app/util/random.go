package util

import (
	"math/rand"
	"strconv"
)

func RandStr(length int) string {
	chars := []rune("ABCDEFGHIGKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxuz1234567890")
	res := make([]rune, length)

	for i := range res {
		res[i] = chars[rand.Intn(len(chars))]
	}

	return string(res)
}

func RandIntStr(n int) string {
	return strconv.Itoa(rand.Intn(n))
}
