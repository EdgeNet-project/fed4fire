package service

import (
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProvision(t *testing.T) {
	s := testService()
	r := testRequest()
	allocateTestSlice(s, r, testRspecMany)
	args := &ProvisionArgs{
		URNs:        []string{testSliceIdentifier.URN()},
		Credentials: []Credential{testSliceCredential},
	}
	reply := &ProvisionReply{}
	err := s.Provision(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
	deployments := listTestDeployments(s)
	assert.Len(t, deployments, 2)
}
