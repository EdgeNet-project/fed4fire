package service

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"
)

func TestDelete_Slice(t *testing.T) {
	s := testService()
	r := testRequest()
	allocateTestSlice(s, r, testRspecMany)
	provisionTestSlice(s, r)
	args := &DeleteArgs{
		URNs:        []string{testSliceIdentifier.URN()},
		Credentials: []Credential{testSliceCredential},
	}
	reply := &DeleteReply{}
	err := s.Delete(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
	assert.Equal(t, constants.GeniStateUnallocated, reply.Data.Value[0].AllocationStatus)
	assert.Equal(t, constants.GeniStateNotReady, reply.Data.Value[0].OperationalStatus)
	slivers := listTestSlivers(s)
	assert.Len(t, slivers, 0)
}
