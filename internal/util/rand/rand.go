package rand

import (
	"crypto/rand"
	"math/big"
)

// Intn 返回 [0,n) 范围内的随机整数；n<=0 时返回 0。
// crypto/rand.Reader 的 Read 按文档保证不返回错误（失败时直接崩溃进程），故 rand.Int 的 err 恒为 nil。
func Intn(n int) int {
	if n <= 0 {
		return 0
	}
	b, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(b.Int64())
}

// letters 是随机串的字符集：数字 + 大小写字母。
const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// String 返回 [0-9a-zA-Z]* 组成、长度为 length 的随机串；length<=0 时返回空串。
// crypto/rand.Read 按文档保证不返回错误（失败时直接崩溃进程），故无需检查 err。
func String(length int) string {
	if length <= 0 {
		return ""
	}

	bytes := make([]byte, length)

	_, _ = rand.Read(bytes)

	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}

	return string(bytes)
}
