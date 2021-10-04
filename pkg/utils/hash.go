package utils

import (
	"crypto/sha256"
	"fmt"
)

func Sha256(s string) string {
	sum := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", sum)
}