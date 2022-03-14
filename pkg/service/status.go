package service

import (
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
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
		Code   Code   `xml:"code"`
		Output string `xml:"output"`
		Value  struct {
			URN     string   `xml:"geni_urn"`
			Slivers []Sliver `xml:"geni_slivers"`
		} `xml:"value"`
	}
}

func (v *StatusReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) error {
	klog.ErrorSDepth(1, err, msg, keysAndValues)
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
	slivers, err := s.AuthorizeAndListSlivers(
		r.Context(),
		r.Header.Get(constants.HttpHeaderUser),
		args.URNs,
		args.Credentials,
	)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorListResources)
	}

	for _, sliver := range slivers {
		allocationStatus, operationalStatus := s.GetSliverStatus(r.Context(), sliver.Name)
		// The spec. says that all the requested slivers belong to the same slice,
		// so it's safe to retrieve the slice URN from any sliver.
		reply.Data.Value.URN = sliver.Spec.SliceURN
		reply.Data.Value.Slivers = append(
			reply.Data.Value.Slivers,
			NewSliver(sliver, allocationStatus, operationalStatus),
		)
	}

	if reply.Data.Value.URN == "" && len(args.URNs) > 0 {
		reply.Data.Value.URN = args.URNs[0]
	}

	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
