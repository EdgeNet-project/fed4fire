package service

import (
	"testing"
)

func TestFindMatchingCredential(t *testing.T) {
	// Untrusted credentials
	_, err := FindCredential(
		testUserIdentifier,
		&testSliceIdentifier,
		[]Credential{testSliceCredential},
		[][]byte{},
	)
	if err == nil {
		t.Errorf("FindCredential() = nil; want error")
	}

	// Missing credentials
	_, err = FindCredential(
		testUserIdentifier,
		&testSliceIdentifier,
		[]Credential{},
		[][]byte{authorityCert},
	)
	if err == nil {
		t.Errorf("FindCredential() = nil; want error")
	}

	// Valid credentials
	_, err = FindCredential(
		testUserIdentifier,
		&testSliceIdentifier,
		[]Credential{testSliceCredential},
		[][]byte{authorityCert},
	)
	if err != nil {
		t.Errorf("FindCredential() = %s; want nil", err)
	}
}
