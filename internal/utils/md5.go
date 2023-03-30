package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 md5
func MD5(data string) string {
	hash := md5.New()
	hash.Write([]byte(data))

	return hex.EncodeToString(hash.Sum(nil))
}
