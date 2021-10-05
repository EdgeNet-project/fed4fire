package service

import (
	"context"
	"encoding/xml"
	"net/http"
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	edgenettestclient "github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned/fake"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	"k8s.io/client-go/kubernetes"
	kubetestclient "k8s.io/client-go/kubernetes/fake"
)

func testService() *Service {
	var edgenetClient versioned.Interface = edgenettestclient.NewSimpleClientset()
	var kubernetesClient kubernetes.Interface = kubetestclient.NewSimpleClientset()
	return &Service{
		ContainerImages: map[string]string{
			"ubuntu2004": "docker.io/library/ubuntu:20.04",
		},
		EdgenetClient:    edgenetClient,
		KubernetesClient: kubernetesClient,
	}
}

func testRequest() *http.Request {
	r, _ := http.NewRequestWithContext(context.TODO(), "", "", nil)
	return r
}

func testNode(name string, ready bool) *v1.Node {
	var readyStatus v1.ConditionStatus = "True"
	if !ready {
		readyStatus = "False"
	}
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"edge-net.io/lat": "s-23.533500",
				"edge-net.io/lon": "w-46.635900",
			},
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   "Ready",
					Status: readyStatus,
				},
			},
		},
	}
}

func TestListResources_BadRspecVersion(t *testing.T) {
	s := testService()
	args := &ListResourcesArgs{Options: ListResourcesOptions{RspecVersion: RspecVersion{
		Type:    "geni",
		Version: "2",
	}}}
	reply := &ListResourcesReply{}
	err := s.ListResources(nil, args, reply)
	if err != nil {
		t.Errorf("GetVersion() = %v; want nil", err)
	}
	got := reply.Data.Code.Code
	if got != geniCodeBadversion {
		t.Errorf("Code = %d; want %d", got, geniCodeBadversion)
	}
}

func TestListResources_MissingRspecVersion(t *testing.T) {
	s := testService()
	args := &ListResourcesArgs{}
	reply := &ListResourcesReply{}
	err := s.ListResources(nil, args, reply)
	if err != nil {
		t.Errorf("GetVersion() = %v; want nil", err)
	}
	got := reply.Data.Code.Code
	if got != geniCodeBadargs {
		t.Errorf("Code = %d; want %d", got, geniCodeBadargs)
	}
}

func TestListResources_NoNodes(t *testing.T) {
	s := testService()
	r := testRequest()
	args := &ListResourcesArgs{
		Options: ListResourcesOptions{
			RspecVersion: RspecVersion{
				Type:    "geni",
				Version: "3",
			}}}
	reply := &ListResourcesReply{}
	err := s.ListResources(r, args, reply)
	if err != nil {
		t.Errorf("GetVersion() = %v; want nil", err)
	}
	if got := reply.Data.Code.Code; got != geniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, geniCodeSuccess)
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
		Options: ListResourcesOptions{
			RspecVersion: RspecVersion{
				Type:    "geni",
				Version: "3",
			}}}
	reply := &ListResourcesReply{}
	err := s.ListResources(r, args, reply)
	if err != nil {
		t.Errorf("GetVersion() = %v; want nil", err)
	}
	if got := reply.Data.Code.Code; got != geniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, geniCodeSuccess)
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
		Options: ListResourcesOptions{
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
	if got := reply.Data.Code.Code; got != geniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, geniCodeSuccess)
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
