package rand

import (
	"crypto/rand"
	"math/big"
)

func Intn(n int) int {
	b, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(b.Int64())
}

const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// String [0-9a-zA-Z]*
func String(length int) string {
	if length <= 0 {
		return ""
	}

	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return ""
	}

	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}

	return string(bytes)
}
