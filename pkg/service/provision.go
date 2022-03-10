package service

import (
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	"net/http"

	"k8s.io/klog/v2"
)

type ProvisionArgs struct {
	URNs        []string
	Credentials []Credential
	Options     ProvisionOptions
}

type ProvisionReply struct {
	Data struct {
		Code  Code `xml:"code"`
		Value struct {
			Rspec   string   `xml:"geni_rspec"`
			Slivers []Sliver `xml:"geni_slivers"`
			Error   string   `xml:"geni_error"`
		} `xml:"value"`
	}
}

type ProvisionOptions struct {
	EndTime      string       `xml:"geni_end_time"`
	RspecVersion RspecVersion `xml:"geni_rspec_version"`
	Users        []struct {
		URN  string   `xml:"urn"`
		Keys []string `xml:"keys"`
	} `xml:"geni_users"`
}

func (v *ProvisionReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Value.Error = fmt.Sprintf("%s: %s", msg, err)
}

// Provision requests that the named geni_allocated slivers be made geni_provisioned,
// instantiating or otherwise realizing the resources, such that they have a valid geni_operational_status
// and may possibly be made geni_ready for experimenter use.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Provision
func (s *Service) Provision(r *http.Request, args *ProvisionArgs, reply *ProvisionReply) error {
	klog.InfoS("%s", args)
	// When only a slice URN is supplied (no specific sliver URNs), this method applies only to the slivers currently in the geni_allocated allocation state.
	// TODO: Provision users here?
	returnRspec := rspec.Rspec{Type: rspec.RspecTypeManifest}
	sliver := Sliver{
		URN:     args.URNs[0],
		Expires: "2022-04-10 14:15:12.237",
		// TODO: Actually store the allocation status in Kube => Sliver CRD?
		AllocationStatus: constants.GeniStateProvisioned,
	}
	//for i, res := range resources {
	//	sliver := Sliver{
	//		URN:              res.Deployment.Annotations[constants.Fed4FireSliver],
	//		Expires:          res.Deployment.Annotations[constants.Fed4FireExpires],
	//		AllocationStatus: constants.GeniStateAllocated,
	//	}
	//	reply.Data.Value.Slivers = append(reply.Data.Value.Slivers, sliver)
	//	returnRspec.Nodes = append(returnRspec.Nodes, requestRspec.Nodes[i])
	//}
	reply.Data.Value.Slivers = append(reply.Data.Value.Slivers, sliver)
	returnRspec.Nodes = append(returnRspec.Nodes, rspec.Node{
		ClientID:           "PC",
		ComponentManagerID: "urn:publicid:IDN+edge-net.org+authority+am",
		Exclusive:          false,
		HardwareType:       rspec.HardwareType{Name: "Test"},
		Services: rspec.Services{
			Logins: []rspec.Login{{
				Authentication: rspec.RspecLoginAuthenticationSSH,
				Hostname:       "example.org",
				Port:           22, // TODO
				Username:       "root",
			}},
		},
		Available: rspec.Available{Now: false},
	})
	// TODO: Return node login info.
	xml_, _ := xml.Marshal(returnRspec)
	//if err != nil {
	//	return reply.SetAndLogError(err, "Failed to serialize response")
	//}
	reply.Data.Value.Rspec = string(xml_)
	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
