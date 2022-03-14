package service

import (
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenew(t *testing.T) {
	s := testService()
	r := testRequest()
	allocateTestSlice(s, r, testRspecMany)
	args := &RenewArgs{
		URNs:           []string{testSliceIdentifier.URN()},
		Credentials:    []Credential{testSliceCredential},
		ExpirationTime: "2100-01-02T15:04:05Z",
	}
	reply := &RenewReply{}
	err := s.Renew(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
	assert.Len(t, reply.Data.Value, 2)
	for _, sliver := range reply.Data.Value {
		assert.Equal(t, args.ExpirationTime, sliver.Expires)
	}
}
