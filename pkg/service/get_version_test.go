package service

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"
)

func TestGetVersion(t *testing.T) {
	s := Service{}
	args := &GetVersionArgs{}
	reply := &GetVersionReply{}
	err := s.GetVersion(nil, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
	assert.Len(t, reply.Data.Value.AdRspecVersions, 1)
	assert.Len(t, reply.Data.Value.RequestRspecVersions, 1)
	assert.Len(t, reply.Data.Value.CredentialTypes, 1)
}
