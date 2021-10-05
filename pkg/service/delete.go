package service

import (
	"fmt"
	"net/http"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"

	"k8s.io/klog/v2"
)

type DeleteArgs struct {
	URNs        []string
	Credentials []Credential
	Options     string // Options
}

type DeleteReply struct {
	Data struct {
		Code  Code     `xml:"code"`
		Value []Sliver `xml:"value"`
	}
}

func (v *DeleteReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = geniCodeError
	// TODO
	// v.Data.Value = fmt.Sprintf("%s: %s", msg, err)
}

// Delete deletes the named slivers, making them geni_unallocated.
// Resources are stopped if necessary, and both de-provisioned and de-allocated.
// No further AM API operations may be performed on slivers that have been deleted.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Delete
func (s *Service) Delete(r *http.Request, args *DeleteArgs, reply *DeleteReply) error {
	//deploymentsClient :=
	// TODO: Check credentials
	// TODO: Check permissions/slice authority
	// TODO: Simplify code by listing all deployments?
	slivers := make([]Sliver, 0)
	for _, urn := range args.URNs {
		identifier, err := identifiers.Parse(urn)
		if err != nil {
			// TODO: Handle error
		}
		//TODO: Handle slice and slivers
		if identifier.ResourceType == identifiers.ResourceTypeSlice {
			subnamespaceName, err := subnamespaceNameForSlice(*identifier)
			if err != nil {
				reply.SetAndLogError(
					err,
					"Failed to build subnamespace name from slice URN",
					"urn",
					urn,
				)
				return nil
			}
			targetNamespace := fmt.Sprintf("%s-%s", s.ParentNamespace, subnamespaceName)
			deploymentsClient := s.KubernetesClient.AppsV1().Deployments(targetNamespace)
			deployments, err := deploymentsClient.List(r.Context(), v1.ListOptions{})
			if err != nil {
				// TODO: Handle error
			}
			for _, deployment := range deployments.Items {
				sliver := Sliver{
					URN:              deployment.Annotations[fed4fireSliver],
					Expires:          deployment.Annotations[fed4fireExpires],
					AllocationStatus: geniStateUnallocated,
				}
				err = deploymentsClient.Delete(r.Context(), deployment.Name, v1.DeleteOptions{})
				if err != nil {
					msg := "Failed to delete deployment"
					klog.ErrorS(err, msg, "name", deployment.Name)
					sliver.Error = fmt.Sprintf("%s: %s", msg, err)
				} else {
					klog.InfoS("Delete deployment", "name", deployment.Name)
				}
				slivers = append(slivers, sliver)
			}
		} else if identifier.ResourceType == identifiers.ResourceTypeSliver {
			// TODO
		} else {
			// TODO: Raise error for invalid resource type.
		}
	}

	reply.Data.Value = slivers
	reply.Data.Code.Code = geniCodeSuccess
	return nil
}
