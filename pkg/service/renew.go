package service

import (
	"fmt"
	"k8s.io/klog/v2"
	"net/http"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RenewArgs struct {
	URNs           []string
	Credentials    []Credential
	ExpirationTime string
	Options        Options
}

type RenewReply struct {
	Data struct {
		Code   Code     `xml:"code"`
		Output string   `xml:"output"`
		Value  []Sliver `xml:"value"`
	}
}

func (v *RenewReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) error {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Output = fmt.Sprintf("%s: %s", msg, err)
	return nil
}

// Renew the named slivers renewed, with their expiration extended.
// If possible, the aggregate should extend the slivers to the requested expiration time,
// or to a sooner time if policy limits apply.
// This method applies to slivers that are geni_allocated or to slivers that are geni_provisioned,
// though different policies may apply to slivers in the different states,
// resulting in much shorter max expiration times for geni_allocated slivers.
func (s *Service) Renew(r *http.Request, args *RenewArgs, reply *RenewReply) error {
	slivers, err := s.AuthorizeAndListSlivers(
		r.Context(),
		r.Header.Get(constants.HttpHeaderUser),
		args.URNs,
		args.Credentials,
	)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorListResources)
	}

	expirationTime, err := time.Parse(time.RFC3339, args.ExpirationTime)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorBadTime)
	}

	for _, sliver := range slivers {
		allocationStatus, operationalStatus := s.GetSliverStatus(r.Context(), sliver.Name)
		if time.Now().After(sliver.Spec.Expires.Time) {
			return reply.SetAndLogError(
				fmt.Errorf("sliver has expired"),
				constants.ErrorUpdateResource,
			)
		}
		sliver.Spec.Expires = metav1.NewTime(expirationTime)
		sliver, err := s.Slivers().Update(r.Context(), &sliver, metav1.UpdateOptions{})
		if err != nil {
			return reply.SetAndLogError(err, constants.ErrorUpdateResource)
		}
		reply.Data.Value = append(
			reply.Data.Value,
			NewSliver(*sliver, allocationStatus, operationalStatus),
		)
	}

	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
