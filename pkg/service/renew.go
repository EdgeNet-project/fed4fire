package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RenewArgs struct {
	URNs           []string
	Credentials    []Credential
	ExpirationTime string
	Options        string
}

type RenewReply struct {
	Data struct {
		Code  Code     `xml:"code"`
		Value []Sliver `xml:"value"`
	}
}

// Renew the named slivers renewed, with their expiration extended.
// If possible, the aggregate should extend the slivers to the requested expiration time,
// or to a sooner time if policy limits apply.
// This method applies to slivers that are geni_allocated or to slivers that are geni_provisioned,
// though different policies may apply to slivers in the different states,
// resulting in much shorter max expiration times for geni_allocated slivers.
func (s *Service) Renew(r *http.Request, args *RenewArgs, reply *RenewReply) error {
	// TODO: Calling Renew on an unknown, deleted or expired sliver (by explicit URN) shall result in an error
	// (e.g. SEARCHFAILED, EXPIRED or ERROR geni_code)
	// (unless geni_best_effort is true, in which case the method may succeed, but return a geni_error for each sliver that failed).
	// TODO: Implement geni_best_effort.
	// It is legal to attempt to renew a sliver to a sooner expiration time than the sliver was previously due to expire.
	expirationTime, err := time.Parse(time.RFC3339, args.ExpirationTime)
	if err != nil {
		// TODO: Handle error
		fmt.Println(err)
	}
	toUpdate := make([]appsv1.Deployment, 0)
	for _, urn := range args.URNs {
		identifier, err := identifiers.Parse(urn)
		if err != nil {
			// TODO: Handle error
			fmt.Println(err)
		}
		deployments, err := s.GetDeployments(r.Context(), *identifier)
		if err != nil {
			// TODO: Handle error
			fmt.Println(err)
		}
		toUpdate = append(toUpdate, deployments...)
	}
	for _, deployment := range toUpdate {
		deployment.Annotations[constants.Fed4FireExpires] = expirationTime.Format(time.RFC3339)
		_, err := s.Deployments().Update(r.Context(), &deployment, metav1.UpdateOptions{})
		if err != nil {
			// TODO: Handle error
			fmt.Println(err)
		}
	}
	// TODO: Return slivers.
	return nil
}
