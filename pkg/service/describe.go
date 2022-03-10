package service

import (
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
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
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Output = fmt.Sprintf("%s: %s", msg, err)
	return nil
}

// Describe retrieves a manifest RSpec describing the resources contained by the named entities,
// e.g. a single slice or a set of the slivers in a slice.
// This listing and description should be sufficiently descriptive to allow experimenters to use the resources.
func (s *Service) Describe(r *http.Request, args *DescribeArgs, reply *DescribeReply) error {
	userIdentifier, err := identifiers.Parse(r.Header.Get(constants.HttpHeaderUser))
	if err != nil {
		return reply.SetAndLogError(err, "Failed to parse user URN")
	}
	resourceIdentifiers, err := identifiers.ParseMultiple(args.URNs)
	if err != nil {
		return reply.SetAndLogError(err, "Failed to parse identifiers")
	}
	_, err = FindMatchingCredentials(*userIdentifier, resourceIdentifiers, args.Credentials, s.TrustedCertificates)
	if err != nil {
		return reply.SetAndLogError(err, "Invalid credentials")
	}
	deployments, err := s.GetDeploymentsMultiple(r.Context(), resourceIdentifiers)
	if err != nil {
		return reply.SetAndLogError(err, "Failed to list deployments")
	}
	for _, deployment := range deployments {
		sliver := Sliver{
			URN:              deployment.Annotations[constants.Fed4FireSliver],
			Expires:          deployment.Annotations[constants.Fed4FireExpires],
			AllocationStatus: constants.GeniStateProvisioned, // TODO: Return the proper state.
		}
		reply.Data.Value.Slivers = append(reply.Data.Value.Slivers, sliver)
	}
	returnRspec := rspec.Rspec{Type: rspec.RspecTypeManifest}
	xml_, _ := xml.Marshal(returnRspec)
	//if err != nil {
	//	return reply.SetAndLogError(err, "Failed to serialize response")
	//}
	reply.Data.Value.Rspec = string(xml_)
	// TODO: Containing slice URN.
	//reply.Data.Value.URN =
	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
