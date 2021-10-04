package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
)

type ListResourcesArgs struct {
	Credentials []Credential
	Options     Options
}

type ListResourcesReply struct {
	Data struct {
		Code  Code   `xml:"code"`
		Value string `xml:"value"`
	}
}

// ListResources returns a listing and description of available resources at this aggregate.
// The resource listing and description provides sufficient information for clients to select among available resources.
// These listings are known as advertisement RSpecs.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#ListResources
func (s *Service) ListResources(r *http.Request, args *ListResourcesArgs, reply *ListResourcesReply) error {
	nodes, err := s.KubernetesClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		reply.Data.Value = "Failed to list nodes"
		reply.Data.Code.Code = geniCodeError
		klog.ErrorS(err, reply.Data.Value)
		return nil
	}

	v := rspec.Rspec{Type: "advertisement"}
	for _, node := range nodes.Items {
		node_ := rspecForNode(node, s.ContainerImages, s.URN)
		if !(args.Options.Available && !node_.Available.Now) {
			v.Nodes = append(v.Nodes, node_)
		}
	}

	xml_, err := xml.Marshal(v)
	if err != nil {
		reply.Data.Value = "Failed to serialize response"
		reply.Data.Code.Code = geniCodeError
		klog.ErrorS(err, reply.Data.Value)
		return nil
	}

	if args.Options.Compressed {
		reply.Data.Value = utils.ZlibBase64(xml_)
	} else {
		reply.Data.Value = string(xml_)
	}

	reply.Data.Code.Code = geniCodeSuccess
	return nil
}

// rspecForNode converts a Kubernetes node to an RSpec node.
func rspecForNode(node corev1.Node, containerImages map[string]string, urn func(resourceType string, resourceName string) string) rspec.Node {
	nodeArch := node.Labels["kubernetes.io/arch"]
	nodeCountry := node.Labels["edge-net.io/country-iso"]
	nodeLatitude := node.Labels["edge-net.io/lat"][1:]
	nodeLongitude := node.Labels["edge-net.io/lon"][1:]
	nodeName := node.Name
	nodeIsReady := true
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" && condition.Status != "True" {
			nodeIsReady = false
			break
		}
	}
	diskImages := make([]rspec.DiskImage, 0)
	for name := range containerImages {
		diskImages = append(diskImages, rspec.DiskImage{
			Name: urn("image", name),
		})
	}
	return rspec.Node{
		ComponentID:        urn("node", nodeName),
		ComponentManagerID: urn("authority", "am"),
		ComponentName:      nodeName,
		Available:          rspec.Available{Now: nodeIsReady},
		Location: rspec.Location{
			Country:   nodeCountry,
			Latitude:  nodeLatitude,
			Longitude: nodeLongitude,
		},
		HardwareType: rspec.HardwareType{
			Name: fmt.Sprintf("kubernetes-%s", nodeArch),
		},
		SliverTypes: []rspec.SliverType{
			{
				Name:       "container",
				DiskImages: diskImages,
			},
		},
	}
}
