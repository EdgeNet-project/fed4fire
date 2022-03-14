package service

import (
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
)

type DescribeArgs struct {
	URNs        []string
	Credentials []Credential
	Options     Options
}

type DescribeReply struct {
	Data struct {
		Code   Code   `xml:"code"`
		Output string `xml:"output"`
		Value  struct {
			Rspec   string   `xml:"geni_rspec"`
			URN     string   `xml:"geni_urn"`
			Slivers []Sliver `xml:"geni_slivers"`
		} `xml:"value"`
	}
}

func (v *DescribeReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) error {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Output = fmt.Sprintf("%s: %s", msg, err)
	return nil
}

// Describe retrieves a manifest RSpec describing the resources contained by the named entities,
// e.g. a single slice or a set of the slivers in a slice.
// This listing and description should be sufficiently descriptive to allow experimenters to use the resources.
func (s *Service) Describe(r *http.Request, args *DescribeArgs, reply *DescribeReply) error {
	slivers, err := s.AuthorizeAndListSlivers(
		r.Context(),
		r.Header.Get(constants.HttpHeaderUser),
		args.URNs,
		args.Credentials,
	)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorListResources)
	}

	returnRspec := rspec.Rspec{Type: rspec.RspecTypeManifest}

	for _, sliver := range slivers {
		logins := make([]rspec.Login, 0)
		service, err := s.Services().Get(r.Context(), sliver.Name, v1.GetOptions{})
		if err != nil {
			return reply.SetAndLogError(err, constants.ErrorGetResource)
		}
		// TODO: Easier way to get the node IP?
		labelSelector := fmt.Sprintf("%s=%s", constants.Fed4FireSliverName, sliver.Name)
		pods, err := s.Pods().List(r.Context(), v1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return reply.SetAndLogError(err, constants.ErrorListResources)
		}
		if len(pods.Items) > 0 {
			nodeName := pods.Items[0].Spec.NodeName
			node, err := s.Nodes().Get(r.Context(), nodeName, v1.GetOptions{})
			if err != nil {
				return reply.SetAndLogError(err, constants.ErrorGetResource)
			}
			for _, address := range node.Status.Addresses {
				if address.Type == corev1.NodeInternalIP {
					logins = append(logins, rspec.Login{
						Authentication: rspec.RspecLoginAuthenticationSSH,
						// We use the IP address of the node here, since some EdgeNet DNS records are broken.
						Hostname: address.Address,
						Port:     int(service.Spec.Ports[0].NodePort),
						Username: "root",
					})
					break
				}
			}
		}
		// The spec. says that all the requested slivers belong to the same slice,
		// so it's safe to retrieve the slice URN from any sliver.
		reply.Data.Value.URN = sliver.Spec.SliceURN
		reply.Data.Value.Slivers = append(reply.Data.Value.Slivers, NewSliver(sliver))
		returnRspec.Nodes = append(returnRspec.Nodes, rspec.Node{
			ComponentManagerID: s.AuthorityIdentifier.URN(),
			Available:          rspec.Available{Now: false},
			ClientID:           sliver.Spec.ClientID,
			SliverID:           sliver.Spec.URN,
			Exclusive:          false,
			HardwareType: rspec.HardwareType{
				// TODO: Set arch if available.
				Name: fmt.Sprintf("kubernetes-unknown"),
			},
			Services: rspec.Services{
				Logins: logins,
			},
		})
	}

	xml_, err := xml.Marshal(returnRspec)
	if err != nil {
		return reply.SetAndLogError(err, "Failed to serialize response")
	}
	reply.Data.Value.Rspec = string(xml_)
	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
