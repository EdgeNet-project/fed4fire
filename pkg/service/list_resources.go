package service

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"

	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
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

func (v *ListResourcesReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = geniCodeError
	v.Data.Value = fmt.Sprintf("%s: %s", msg, err)
}

// ListResources returns a listing and description of available resources at this aggregate.
// The resource listing and description provides sufficient information for clients to select among available resources.
// These listings are known as advertisement RSpecs.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#ListResources
func (s *Service) ListResources(
	r *http.Request,
	args *ListResourcesArgs,
	reply *ListResourcesReply,
) error {
	if strings.ToLower(args.Options.RspecVersion.Type) != "geni" || args.Options.RspecVersion.Version != "3" {
		reply.Data.Code.Code = geniCodeBadversion
		return nil
	}

	nodes, err := s.KubernetesClient.CoreV1().Nodes().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		reply.SetAndLogError(err, "Failed to list nodes")
		return nil
	}

	v := rspec.Rspec{Type: "advertisement"}
	for _, node := range nodes.Items {
		node_ := rspecForNode(node, s.AuthorityIdentifier, s.ContainerImages)
		if !(args.Options.Available && !node_.Available.Now) {
			v.Nodes = append(v.Nodes, node_)
		}
	}

	xml_, err := xml.Marshal(v)
	if err != nil {
		reply.SetAndLogError(err, "Failed to serialize response")
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
func rspecForNode(
	node corev1.Node,
	authorityIdentifier identifiers.Identifier,
	containerImages map[string]string,
) rspec.Node {
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
			Name: authorityIdentifier.Copy("image", name).URN(),
		})
	}
	return rspec.Node{
		ComponentID:        authorityIdentifier.Copy("node", nodeName).URN(),
		ComponentManagerID: authorityIdentifier.URN(),
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
