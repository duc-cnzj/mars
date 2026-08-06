package hasher

import (
	"crypto/sha256"
	"fmt"
)

// Hash 对 data 计算 sha256 并以十六进制字符串返回。
func Hash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))

	return fmt.Sprintf("%x", h.Sum(nil))
}
