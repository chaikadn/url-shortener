package util

import "math/rand"

func RandStr(length int) string {
	chars := []rune("ABCDEFGHIGKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxuz1234567890")
	res := make([]rune, length)

	for i := range res {
		res[i] = chars[rand.Intn(len(chars))]
	}

	return string(res)
}
