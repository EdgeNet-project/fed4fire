package service

import (
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"k8s.io/klog/v2"
	"net/http"
)

type PerformOperationalActionArgs struct {
	URNs        []string
	Credentials []Credential
	Action      string
	Options     Options
}

type PerformOperationalActionReply struct {
	Data struct {
		Code   Code     `xml:"code"`
		Output string   `xml:"output"`
		Value  []Sliver `xml:"value"`
	}
}

func (v *PerformOperationalActionReply) SetAndLogError(
	err error,
	msg string,
	keysAndValues ...interface{},
) error {
	klog.ErrorSDepth(1, err, msg, keysAndValues)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Output = fmt.Sprintf("%s: %s", msg, err)
	return nil
}

// PerformOperationalAction performs the named operational action on the named slivers,
// possibly changing the geni_operational_status of the named slivers, e.g. 'start' a VM.
// For valid operations and expected states, consult the state diagram advertised in the aggregate's advertisement RSpec.
func (s *Service) PerformOperationalAction(
	r *http.Request,
	args *PerformOperationalActionArgs,
	reply *PerformOperationalActionReply,
) error {
	slivers, err := s.AuthorizeAndListSlivers(
		r.Context(),
		r.Header.Get(constants.HttpHeaderUser),
		args.URNs,
		args.Credentials,
	)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorListResources)
	}

	if args.Action != "geni_start" {
		return reply.SetAndLogError(
			fmt.Errorf("action must be geni_start"),
			constants.ErrorBadAction,
		)
	}

	// Do nothing, `geni_start` is a no-op for us.

	for _, sliver := range slivers {
		allocationStatus, operationalStatus := s.GetSliverStatus(r.Context(), sliver.Name)
		reply.Data.Value = append(
			reply.Data.Value,
			NewSliver(sliver, allocationStatus, operationalStatus),
		)
	}
	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
