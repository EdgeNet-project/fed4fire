package service

import (
	"fmt"
	"net/http"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"

	"k8s.io/klog/v2"
)

type DeleteArgs struct {
	URNs        []string
	Credentials []Credential
	Options     Options
}

type DeleteReply struct {
	Data struct {
		Code  Code     `xml:"code"`
		Value []Sliver `xml:"value"`
	}
}

func (v *DeleteReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) error {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Value = []Sliver{{
		Error: fmt.Sprintf("%s: %s", msg, err),
	}}
	return nil
}

// Delete deletes the named slivers, making them geni_unallocated.
// Resources are stopped if necessary, and both de-provisioned and de-allocated.
// No further AM API operations may be performed on slivers that have been deleted.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Delete
func (s *Service) Delete(r *http.Request, args *DeleteArgs, reply *DeleteReply) error {
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
			AllocationStatus: constants.GeniStateUnallocated,
		}
		err := s.Deployments().Delete(r.Context(), deployment.Name, metav1.DeleteOptions{})
		if err != nil {
			return reply.SetAndLogError(err, "Failed to delete deployment", "name", deployment.Name)
		}
		klog.InfoS("Deleted deployment", "name", deployment.Name)
		reply.Data.Value = append(reply.Data.Value, sliver)
	}
	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
