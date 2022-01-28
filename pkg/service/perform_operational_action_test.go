package service

import (
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"testing"
)

func TestPerformOperationAction(t *testing.T) {
	s := testService()
	r := testRequest()

	// Allocate one node
	rs := `
<rspec type="request" generated="2013-01-16T14:20:39Z" xsi:schemaLocation="http://www.geni.net/resources/rspec/3 http://www.geni.net/resources/rspec/3/request.xsd " xmlns:client="http://www.protogeni.net/resources/rspec/ext/client/1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns="http://www.geni.net/resources/rspec/3">
  <node client_id="PC1" component_manager_id="urn:publicid:IDN+ilabt.imec.be+authority+cm" exclusive="false">
    <sliver_type name="container"/>
  </node>
</rspec>
`
	args := &AllocateArgs{
		SliceURN:    testSliceIdentifier.URN(),
		Credentials: []Credential{testSliceCredential},
		Rspec:       rs,
	}
	reply := &AllocateReply{}
	err := s.Allocate(r, args, reply)
	if err != nil {
		t.Errorf("Allocate() = %v; want nil", err)
	}
	got := reply.Data.Code.Code
	if got != constants.GeniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, constants.GeniCodeSuccess)
	}

	// Update the node SSH keys
	args2 := &PerformOperationalActionArgs{
		URNs:        []string{testSliceIdentifier.URN()},
		Credentials: []Credential{testSliceCredential},
		Action:      "geni_update_users",
		Options: UpdateUsersOptions{
			Users: []struct {
				URN  string   `xml:"urn"`
				Keys []string `xml:"keys"`
			}{{URN: "", Keys: []string{"some-ssh-key"}}},
		},
	}
	reply2 := &PerformOperationalActionReply{}
	err = s.PerformOperationalAction(r, args2, reply2)
	if err != nil {
		t.Errorf("PerformOperationalAction() = %v; want nil", err)
	}
	got = reply.Data.Code.Code
	if got != constants.GeniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, constants.GeniCodeSuccess)
	}
	// TODO: Test that the key is present in the configmap.
}
