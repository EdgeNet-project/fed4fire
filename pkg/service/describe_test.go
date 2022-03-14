package service

import (
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDescribe(t *testing.T) {
	s := testService()
	r := testRequest()
	allocateTestSlice(s, r, testRspecMany)
	args := &DescribeArgs{
		URNs:        []string{testSliceIdentifier.URN()},
		Credentials: []Credential{testSliceCredential},
	}
	reply := &DescribeReply{}
	err := s.Describe(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
	assert.Len(t, reply.Data.Value.Slivers, 2)
}
