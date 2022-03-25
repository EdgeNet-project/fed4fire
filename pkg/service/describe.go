package service

import (
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
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
	klog.ErrorSDepth(1, err, msg, keysAndValues)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Output = fmt.Sprintf("%s: %s", msg, err)
	return nil
}

// Describe retrieves a manifest RSpec describing the resources contained by the named entities,
// e.g. a single slice or a set of the slivers in a slice.
// This listing and description should be sufficiently descriptive to allow experimenters to use the resources.
func (s *Service) Describe(r *http.Request, args *DescribeArgs, reply *DescribeReply) error {
	slivers, err := s.AuthorizeAndListSlivers(r, args.URNs, args.Credentials)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorListResources)
	}

	returnRspec := rspec.Rspec{Type: rspec.RspecTypeManifest}

	for _, sliver := range slivers {
		available := rspec.Available{Now: false}
		var hardwareType *rspec.HardwareType
		var services *rspec.Services
		arch, host, port := s.GetSliverArchHostPort(r.Context(), sliver.Name)
		if arch != nil && host != nil && port != nil {
			available.Now = true
			services = &rspec.Services{
				Logins: []rspec.Login{{
					Authentication: rspec.RspecLoginAuthenticationSSH,
					Hostname:       *host,
					Port:           *port,
					Username:       "root",
				}},
			}
		}
		returnRspec.Nodes = append(returnRspec.Nodes, rspec.Node{
			// TODO: Node component ID / name
			ComponentManagerID: s.AuthorityIdentifier.URN(),
			ClientID:           sliver.Spec.ClientID,
			SliverID:           sliver.Spec.URN,
			Available:          &available,
			Exclusive:          false,
			HardwareType:       hardwareType,
			Services:           services,
			SliverType: rspec.SliverType{
				Name: "container",
			},
		})
		allocationStatus, operationalStatus := s.GetSliverStatus(r.Context(), sliver.Name)
		// The spec. says that all the requested slivers belong to the same slice,
		// so it's safe to retrieve the slice URN from any sliver.
		reply.Data.Value.URN = sliver.Spec.SliceURN
		reply.Data.Value.Slivers = append(
			reply.Data.Value.Slivers,
			NewSliver(sliver, allocationStatus, operationalStatus),
		)
	}

	xml_, err := MarshalRspec(returnRspec, args.Options.Compressed)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorSerializeRspec)
	}
	reply.Data.Value.Rspec = xml_
	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
