package rand

import (
	"crypto/rand"
	"math/big"
)

func Intn(n int) int {
	b, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(b.Int64())
}
