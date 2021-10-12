package service

import (
	"context"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	"net/http"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	edgenettestclient "github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned/fake"
	"k8s.io/client-go/kubernetes"
	kubetestclient "k8s.io/client-go/kubernetes/fake"
)

const testSliceUrn = "urn:publicid:IDN+example.org+slice+test"
const testUserUrn = "urn:publicid:IDN+example.org+user+test"

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
	r.Header.Set(utils.HttpHeaderUser, testUserUrn)
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
