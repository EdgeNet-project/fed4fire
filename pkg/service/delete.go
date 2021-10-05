package service

import (
	"net/http"

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
	// TODO: Check credentials
	//slivers := make([]Sliver, 0)
	//for _, urn := range args.URNs {
	//	TODO: Handle slice and slivers
	//}
	return nil
}
