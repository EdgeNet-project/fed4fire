package service

import (
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog/v2"
)

type DeleteArgs struct {
	URNs        []string
	Credentials []Credential
	Options     Options
}

type DeleteReply struct {
	Data struct {
		Code   Code     `xml:"code"`
		Output string   `xml:"output"`
		Value  []Sliver `xml:"value"`
	}
}

func (v *DeleteReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) error {
	klog.ErrorSDepth(1, err, msg, keysAndValues)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Output = fmt.Sprintf("%s: %s", msg, err)
	return nil
}

// Delete deletes the named slivers, making them geni_unallocated.
// Resources are stopped if necessary, and both de-provisioned and de-allocated.
// No further AM API operations may be performed on slivers that have been deleted.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Delete
func (s *Service) Delete(r *http.Request, args *DeleteArgs, reply *DeleteReply) error {
	slivers, err := s.AuthorizeAndListSlivers(r, args.URNs, args.Credentials)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorListResources)
	}

	for _, sliver := range slivers {
		err := s.Slivers().Delete(r.Context(), sliver.Name, metav1.DeleteOptions{})
		if err != nil {
			return reply.SetAndLogError(err, constants.ErrorDeleteResource)
		}
		allocationStatus, operationalStatus := s.GetSliverStatus(r.Context(), sliver.Name)
		reply.Data.Value = append(
			reply.Data.Value,
			NewSliver(sliver, allocationStatus, operationalStatus),
		)
	}

	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
