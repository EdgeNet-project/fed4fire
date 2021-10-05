package service

import (
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

type ListResourcesOptions struct {
	// XML-RPC boolean value indicating whether the caller is interested in
	// all resources or available resources.
	// If this value is true (1), the result should contain only available resources.
	// If this value is false (0) or unspecified, both available and allocated resources should be returned.
	// The Aggregate Manager is free to limit visibility of certain resources based on the credentials parameter.
	Available bool `xml:"geni_available"`
	// XML-RPC boolean value indicating whether the caller would like the result to be compressed.
	// If the value is true (1), the returned resource list will be compressed according to RFC 1950.
	// If the value is false (0) or unspecified, the return will be text.
	Compressed bool `xml:"geni_compressed"`
	// Requested expiration of all new slivers, may be ignored by aggregates.
	EndTime string `xml:"geni_end_time"`
	// XML-RPC struct indicating the type and version of Advertisement RSpec to return.
	// The struct contains 2 members, type and version. type and version are case-insensitive strings,
	// matching those in geni_ad_rspec_versions as returned by GetVersion at this aggregate.
	// This option is required, and aggregates are expected to return a geni_code of 1 (BADARGS) if it is missing.
	// Aggregates should return a geni_code of 4 (BADVERSION) if the requested RSpec version
	// is not one advertised as supported in GetVersion.
	RspecVersion struct {
		Type    string `xml:"type"`
		Version string `xml:"version"`
	} `xml:"geni_rspec_version"`
}

type ListResourcesArgs struct {
	Credentials []Credential
	Options     ListResourcesOptions
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
