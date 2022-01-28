package service

import (
	"testing"
)

func TestFindMatchingCredential(t *testing.T) {
	// Untrusted credentials
	_, err := FindMatchingCredential(
		testUserIdentifier,
		testSliceIdentifier,
		[]Credential{testSliceCredential},
		[][]byte{},
	)
	if err == nil {
		t.Errorf("FindMatchingCredential() = nil; want error")
	}

	// Missing credentials
	_, err = FindMatchingCredential(
		testUserIdentifier,
		testSliceIdentifier,
		[]Credential{},
		[][]byte{authorityCert},
	)
	if err == nil {
		t.Errorf("FindMatchingCredential() = nil; want error")
	}

	// Valid credentials
	_, err = FindMatchingCredential(
		testUserIdentifier,
		testSliceIdentifier,
		[]Credential{testSliceCredential},
		[][]byte{authorityCert},
	)
	if err != nil {
		t.Errorf("FindMatchingCredential() = %s; want nil", err)
	}
}
