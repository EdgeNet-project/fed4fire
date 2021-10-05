package service

import (
	"context"
	"encoding/xml"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	"net/http"
	"testing"

	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	edgenettestclient "github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned/fake"
	"k8s.io/client-go/kubernetes"
	kubetestclient "k8s.io/client-go/kubernetes/fake"
)

func TestListResources_BadRspecVersion(t *testing.T) {
	s := Service{}
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
	s := Service{}
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
	var edgenetClient versioned.Interface = edgenettestclient.NewSimpleClientset()
	var kubernetesClient kubernetes.Interface = kubetestclient.NewSimpleClientset()
	r, _ := http.NewRequestWithContext(context.TODO(), "", "", nil)
	s := Service{
		EdgenetClient:    edgenetClient,
		KubernetesClient: kubernetesClient,
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
	if got := len(v.Nodes); got != 0 {
		t.Errorf("len(Nodes) = %d; want %d", got, 0)
	}
}
