package service

import (
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPerformOperationalAction_BadAction(t *testing.T) {
	s := testService()
	r := testRequest()
	allocateTestSlice(s, r, testRspecMany)
	args := &PerformOperationalActionArgs{
		URNs:        []string{testSliceIdentifier.URN()},
		Credentials: []Credential{testSliceCredential},
		Action:      "invalid",
	}
	reply := &PerformOperationalActionReply{}
	err := s.PerformOperationalAction(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeError, reply.Data.Code.Code)
}

func TestPerformOperationalAction(t *testing.T) {
	s := testService()
	r := testRequest()
	allocateTestSlice(s, r, testRspecMany)
	args := &PerformOperationalActionArgs{
		URNs:        []string{testSliceIdentifier.URN()},
		Credentials: []Credential{testSliceCredential},
		Action:      constants.GeniActionStart,
	}
	reply := &PerformOperationalActionReply{}
	err := s.PerformOperationalAction(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
}
