package service

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"
)

func TestAllocate_Single(t *testing.T) {
	s := testService()
	r := testRequest()
	args := &AllocateArgs{
		SliceURN:    testSliceIdentifier.URN(),
		Credentials: []Credential{testSliceCredential},
		Rspec:       testRspecSingle,
	}
	reply := &AllocateReply{}
	err := s.Allocate(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
	slivers := listTestSlivers(s)
	assert.Len(t, slivers, 1)
}

func TestAllocate_Many(t *testing.T) {
	s := testService()
	r := testRequest()

	// First request with one node
	args := &AllocateArgs{
		SliceURN:    testSliceIdentifier.URN(),
		Credentials: []Credential{testSliceCredential},
		Rspec:       testRspecSingle,
	}
	reply := &AllocateReply{}
	err := s.Allocate(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)

	// Second request with the first node repeated and a new node
	args = &AllocateArgs{
		SliceURN:    testSliceIdentifier.URN(),
		Credentials: []Credential{testSliceCredential},
		Rspec:       testRspecMany,
	}
	reply = &AllocateReply{}
	err = s.Allocate(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
	slivers := listTestSlivers(s)
	assert.Len(t, slivers, 2)
}
