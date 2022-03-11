// Package naming maps GENI identifiers to Kubernetes-compatible names.
// The current strategy is to use the first 8 bytes of a SHA512 hash represented as a hexadecimal string.
package naming

import (
	"crypto/sha512"
	"fmt"
)

func SliceHash(sliceUrn string) string {
	return "h" + sha512Sum(sliceUrn)[:16]
}

func SliverName(sliceUrn string, clientId string) string {
	return "h" + sha512Sum(sliceUrn + clientId)[:16]
}

func sha512Sum(s string) string {
	h := sha512.Sum512([]byte(s))
	return fmt.Sprintf("%x", h)
}
