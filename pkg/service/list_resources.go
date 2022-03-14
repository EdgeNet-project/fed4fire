package service

import (
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"net/http"

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
		Code   Code   `xml:"code"`
		Output string `xml:"output"`
		Value  string `xml:"value"`
	}
}

func (v *ListResourcesReply) SetAndLogError(
	err error,
	msg string,
	code int,
	keysAndValues ...interface{},
) error {
	klog.ErrorSDepth(1, err, msg, keysAndValues)
	v.Data.Code.Code = code
	v.Data.Output = fmt.Sprintf("%s: %s", msg, err)
	return nil
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
	userIdentifier, err := identifiers.Parse(r.Header.Get(constants.HttpHeaderUser))
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorBadIdentifier, constants.GeniCodeError)
	}
	_, err = FindCredential(*userIdentifier, nil, args.Credentials, s.TrustedCertificates)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorBadCredentials, constants.GeniCodeBadargs)
	}

	nodes, err := s.Nodes().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorListResources, constants.GeniCodeError)
	}

	v := rspec.Rspec{Type: rspec.RspecTypeAdvertisement}
	for _, node := range nodes.Items {
		node_ := rspecForNode(node, s.AuthorityIdentifier, s.ContainerImages)
		if !(args.Options.Available && !node_.Available.Now) {
			v.Nodes = append(v.Nodes, node_)
		}
	}

	xml_, err := xml.Marshal(v)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorSerializeRspec, constants.GeniCodeError)
	}

	if args.Options.Compressed {
		reply.Data.Value = utils.CompressZlibBase64(xml_)
	} else {
		reply.Data.Value = string(xml_)
	}

	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}

// rspecForNode converts a Kubernetes node to an RSpec node.
func rspecForNode(
	node corev1.Node,
	authorityIdentifier identifiers.Identifier,
	containerImages map[string]string,
) rspec.Node {
	nodeArch := node.Labels[corev1.LabelArchStable]
	nodeCountry := node.Labels[constants.EdgeNetLabelCountryISO]
	nodeLatitude := node.Labels[constants.EdgeNetLabelLatitude]
	nodeLongitude := node.Labels[constants.EdgeNetLabelLongitude]
	// n39.92050 -> 39.92050
	if len(nodeLatitude) > 1 {
		nodeLatitude = nodeLatitude[1:]
	}
	if len(nodeLongitude) > 1 {
		nodeLongitude = nodeLongitude[1:]
	}
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
			Name: authorityIdentifier.Copy(identifiers.ResourceTypeImage, name).URN(),
		})
	}
	return rspec.Node{
		ComponentID:        authorityIdentifier.Copy(identifiers.ResourceTypeNode, nodeName).URN(),
		ComponentManagerID: authorityIdentifier.URN(),
		ComponentName:      nodeName,
		Available:          rspec.Available{Now: nodeIsReady},
		Location: rspec.Location{
			Country:   nodeCountry,
			Latitude:  nodeLatitude,
			Longitude: nodeLongitude,
		},
		HardwareType: rspec.HardwareType{
			Name: fmt.Sprintf(nodeArch),
		},
		SliverType: rspec.SliverType{
			Name:       "container",
			DiskImages: diskImages,
		},
	}
}
