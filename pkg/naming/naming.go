// Package naming maps GENI identifiers to Kubernetes-compatible names.
// The current strategy is to use the first 8 bytes of a SHA512 hash represented as a hexadecimal string.
package naming

import (
	"crypto/sha512"
	"fmt"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
)

func SliceHash(sliceIdentifier identifiers.Identifier) string {
	return sha512Sum(sliceIdentifier.URN())[:16]
}

func SliverName(sliceIdentifier identifiers.Identifier, clientId string) (string, error) {
	if sliceIdentifier.ResourceType != identifiers.ResourceTypeSlice {
		return "", fmt.Errorf("URN resource type must be `slice`")
	}
	s := "f4f-" + sha512Sum(sliceIdentifier.URN() + clientId)[:16]
	return s, nil
}

func sha512Sum(s string) string {
	h := sha512.Sum512([]byte(s))
	return fmt.Sprintf("%x", h)
}
