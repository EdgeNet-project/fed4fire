package service

import "testing"

func TestFindMatchingCredential(t *testing.T) {
	_, err := FindMatchingCredential(
		testUserIdentifier,
		testSliceIdentifier,
		[]Credential{testSliceCredential},
		[][]byte{},
	)
	if err == nil {
		t.Errorf("FindMatchingCredential() = nil; want error")
	}
	_, err = FindMatchingCredential(
		testUserIdentifier,
		testSliceIdentifier,
		[]Credential{},
		[][]byte{authorityCert},
	)
	if err == nil {
		t.Errorf("FindMatchingCredential() = nil; want error")
	}
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
