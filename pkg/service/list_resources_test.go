package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
)

func TestListResources_NoNodes(t *testing.T) {
	s := testService()
	r := testRequest()
	args := &ListResourcesArgs{
		Credentials: []Credential{testSliceCredential},
		Options: Options{
			RspecVersion: RspecVersion{
				Type:    "geni",
				Version: "3",
			}}}
	reply := &ListResourcesReply{}
	err := s.ListResources(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
	v := unmarshalTestRspec(reply.Data.Value)
	assert.Equal(t, rspec.RspecTypeAdvertisement, v.Type)
	assert.Len(t, v.Nodes, 0)
}

func TestListResources_Nodes(t *testing.T) {
	s := testService()
	r := testRequest()
	nodes := []*v1.Node{
		testNode("node-1", true),
		testNode("node-2", true),
		testNode("node-3", false),
	}
	for _, node := range nodes {
		s.KubernetesClient.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{})
	}
	args := &ListResourcesArgs{
		Credentials: []Credential{testSliceCredential},
		Options: Options{
			RspecVersion: RspecVersion{
				Type:    "geni",
				Version: "3",
			}}}
	reply := &ListResourcesReply{}
	err := s.ListResources(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
	v := unmarshalTestRspec(reply.Data.Value)
	assert.Equal(t, rspec.RspecTypeAdvertisement, v.Type)
	assert.Len(t, v.Nodes, 3)
}

func TestListResources_NodesAvailableCompressed(t *testing.T) {
	s := testService()
	r := testRequest()
	nodes := []*v1.Node{
		testNode("node-1", true),
		testNode("node-2", true),
		testNode("node-3", false),
	}
	for _, node := range nodes {
		s.KubernetesClient.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{})
	}
	args := &ListResourcesArgs{
		Credentials: []Credential{testSliceCredential},
		Options: Options{
			Available:  true,
			Compressed: true,
			RspecVersion: RspecVersion{
				Type:    "geni",
				Version: "3",
			}}}
	reply := &ListResourcesReply{}
	err := s.ListResources(r, args, reply)
	assert.Nil(t, err)
	assert.Equal(t, constants.GeniCodeSuccess, reply.Data.Code.Code)
	v := unmarshalTestRspec(reply.Data.Value)
	assert.Equal(t, rspec.RspecTypeAdvertisement, v.Type)
	assert.Len(t, v.Nodes, 2)
}
