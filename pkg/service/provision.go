package service

import (
	"net/http"

	"k8s.io/klog/v2"
)

type ProvisionArgs struct {
	SliceURN    string
	Credentials []Credential
	Rspec       string
	Options     string // Options
}

type ProvisionReply struct {
	Data struct {
		Code  Code `xml:"code"`
		Value struct {
			Rspec   string   `xml:"geni_rspec"`
			Slivers []Sliver `xml:"geni_slivers"`
		} `xml:"value"`
	}
}

func (v *ProvisionReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = geniCodeError
	// TODO
	// v.Data.Value = fmt.Sprintf("%s: %s", msg, err)
}

// Provision requests that the named geni_allocated slivers be made geni_provisioned,
// instantiating or otherwise realizing the resources, such that they have a valid geni_operational_status
// and may possibly be made geni_ready for experimenter use.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Provision
func (s *Service) Provision(r *http.Request, args *ProvisionArgs, reply *ProvisionReply) error {
	// When only a slice URN is supplied (no specific sliver URNs), this method applies only to the slivers currently in the geni_allocated allocation state.
	return nil
}
