package service

import (
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"
)

func TestAllocate_Single(t *testing.T) {
	s := testService()
	r := testRequest()
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
}

func TestAllocate_Many(t *testing.T) {
	s := testService()
	r := testRequest()
	// First request with one node
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
	// Second request with the first node repeated and a new node
	rs = `
<rspec type="request" generated="2013-01-16T14:20:39Z" xsi:schemaLocation="http://www.geni.net/resources/rspec/3 http://www.geni.net/resources/rspec/3/request.xsd " xmlns:client="http://www.protogeni.net/resources/rspec/ext/client/1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns="http://www.geni.net/resources/rspec/3">
  <node client_id="PC1" component_manager_id="urn:publicid:IDN+ilabt.imec.be+authority+cm" exclusive="false">
    <sliver_type name="container"/>
  </node>
  <node client_id="PC2" component_manager_id="urn:publicid:IDN+ilabt.imec.be+authority+cm" exclusive="false">
    <sliver_type name="container"/>
  </node>
</rspec>
`
	args = &AllocateArgs{
		SliceURN:    testSliceIdentifier.URN(),
		Credentials: []Credential{testSliceCredential},
		Rspec:       rs,
	}
	reply = &AllocateReply{}
	err = s.Allocate(r, args, reply)
	if err != nil {
		t.Errorf("Allocate() = %v; want nil", err)
	}
	got = reply.Data.Code.Code
	if got != constants.GeniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, constants.GeniCodeSuccess)
	}
}
