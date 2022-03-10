package service

import (
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	"k8s.io/klog/v2"
	"net/http"
)

type StatusArgs struct {
	URNs        []string
	Credentials []Credential
	Options     Options
}

type StatusReply struct {
	Data struct {
		Code Code `xml:"code"`
		// TODO: Check in other parts of the code where we need `Output` instead of `Error`.
		Output string `xml:"output"`
		Value  struct {
			URN     string   `xml:"geni_urn"`
			Slivers []Sliver `xml:"geni_slivers"`
		} `xml:"value"`
	}
}

func (v *StatusReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) error {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Output = fmt.Sprintf("%s: %s", msg, err)
	return nil
}

// Status gets the status of a sliver or slivers belonging to a single slice at the given aggregate.
// Status may include other dynamic reservation or instantiation information as required by the resource type and aggregate.
// This method is used to provide updates on the state of the resources after the completion of Provision,
// which began to asynchronously provision the resources. This should be relatively dynamic data,
// not descriptive data as returned in the manifest RSpec.
func (s *Service) Status(r *http.Request, args *StatusArgs, reply *StatusReply) error {
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
	// TODO: Enclosing slice URN / check that all the slivers belong to the same slice OR a single slice identifier.
	reply.Data.Value.URN = args.URNs[0]
	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
