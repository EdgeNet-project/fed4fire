package service

import (
	"fmt"
	"net/http"

	"github.com/EdgeNet-project/fed4fire/pkg/naming"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	// Delete moves 1 or more slivers from either state 2 or 3 (geni_allocated or geni_provisioned),
	// back to state 1 (geni_unallocated).
	// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3/CommonConcepts#SliverAllocationStates
	// TODO: Check credentials
	// TODO: Check permissions/slice authority
	// TODO: Delete configmaps
	deploymentsToDelete := make([]appsv1.Deployment, 0)
	servicesToDelete := make([]corev1.Service, 0)
	slivers := make([]Sliver, 0)
	for _, urn := range args.URNs {
		identifier, err := identifiers.Parse(urn)
		if err != nil {
			// TODO: Handle error
			fmt.Println(err)
		}
		if identifier.ResourceType == identifiers.ResourceTypeSlice {
			sliceHash := naming.SliceHash(*identifier)
			labelSelector := fmt.Sprintf("%s=%s", fed4fireSliceHash, sliceHash)
			deployments, err := s.Deployments().List(r.Context(), metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err != nil {
				// TODO: Handle error
				fmt.Println(err)
			}
			services, err := s.Services().List(r.Context(), metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err != nil {
				// TODO: Handle error
				fmt.Println(err)
			}
			deploymentsToDelete = append(deploymentsToDelete, deployments.Items...)
			servicesToDelete = append(servicesToDelete, services.Items...)
		} else if identifier.ResourceType == identifiers.ResourceTypeSliver {
			deployment, err := s.Deployments().Get(r.Context(), identifier.ResourceName, metav1.GetOptions{})
			if err != nil {
				// TODO: Handle error
				fmt.Println(err)
			}
			service, err := s.Services().Get(r.Context(), identifier.ResourceName, metav1.GetOptions{})
			if err != nil {
				// TODO: Handle error
				fmt.Println(err)
			}
			deploymentsToDelete = append(deploymentsToDelete, *deployment)
			servicesToDelete = append(servicesToDelete, *service)
		} else {
			// TODO: Raise error for invalid resource type.
		}
	}

	for _, deployment := range deploymentsToDelete {
		sliver := Sliver{
			URN:              deployment.Annotations[fed4fireSliver],
			Expires:          deployment.Annotations[fed4fireExpires],
			AllocationStatus: geniStateUnallocated,
		}
		err := s.Deployments().Delete(r.Context(), deployment.Name, metav1.DeleteOptions{})
		if err != nil {
			msg := "Failed to delete deployment"
			klog.ErrorS(err, msg, "name", deployment.Name)
			sliver.Error = fmt.Sprintf("%s: %s", msg, err)
		} else {
			klog.InfoS("Deleted deployment", "name", deployment.Name)
		}
		slivers = append(slivers, sliver)
	}

	for _, service := range servicesToDelete {
		err := s.Services().Delete(r.Context(), service.Name, metav1.DeleteOptions{})
		if err != nil {
			msg := "Failed to delete service"
			klog.ErrorS(err, msg, "name", service.Name)
		} else {
			klog.InfoS("Deleted service", "name", service.Name)
		}
	}

	reply.Data.Value = slivers
	reply.Data.Code.Code = geniCodeSuccess
	return nil
}
