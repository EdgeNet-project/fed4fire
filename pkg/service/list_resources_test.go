package service

import (
	"context"
	"encoding/xml"
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"

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
	if err != nil {
		t.Errorf("GetVersion() = %v; want nil", err)
	}
	if got := reply.Data.Code.Code; got != constants.GeniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, constants.GeniCodeSuccess)
	}
	v := rspec.Rspec{}
	_ = xml.Unmarshal([]byte(reply.Data.Value), &v)
	if got := v.Type; got != rspec.RspecTypeAdvertisement {
		t.Errorf("Type = %s; want %s", got, rspec.RspecTypeAdvertisement)
	}
	if got := len(v.Nodes); got != 0 {
		t.Errorf("len(Nodes) = %d; want %d", got, 0)
	}
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
	if err != nil {
		t.Errorf("GetVersion() = %v; want nil", err)
	}
	if got := reply.Data.Code.Code; got != constants.GeniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, constants.GeniCodeSuccess)
	}
	v := rspec.Rspec{}
	_ = xml.Unmarshal([]byte(reply.Data.Value), &v)
	if got := v.Type; got != rspec.RspecTypeAdvertisement {
		t.Errorf("Type = %s; want %s", got, rspec.RspecTypeAdvertisement)
	}
	if got := len(v.Nodes); got != 3 {
		t.Errorf("len(Nodes) = %d; want %d", got, 3)
	}
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
	if err != nil {
		t.Errorf("GetVersion() = %v; want nil", err)
	}
	if got := reply.Data.Code.Code; got != constants.GeniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, constants.GeniCodeSuccess)
	}
	v := rspec.Rspec{}
	_ = xml.Unmarshal(utils.DecompressZlibBase64(reply.Data.Value), &v)
	if got := v.Type; got != rspec.RspecTypeAdvertisement {
		t.Errorf("Type = %s; want %s", got, rspec.RspecTypeAdvertisement)
	}
	if got := len(v.Nodes); got != 2 {
		t.Errorf("len(Nodes) = %d; want %d", got, 2)
	}
}
