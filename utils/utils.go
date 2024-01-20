package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// StitchingString 多个字符串进行拼接
func StitchingString(args ...string) string {
	var Builder strings.Builder

	for i := 0; i < len(args); i++ {
		Builder.WriteString(args[i])
	}

	return Builder.String()
}
func HashSha256(msg []byte) string {
	var h = sha256.New()
	h.Write(msg)
	return hex.EncodeToString(h.Sum(nil))
}
