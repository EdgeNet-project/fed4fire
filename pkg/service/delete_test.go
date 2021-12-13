package service

import (
	"context"
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDelete_Slice(t *testing.T) {
	s := testService()
	r := testRequest()
	rs := `
<rspec type="request" generated="2013-01-16T14:20:39Z" xsi:schemaLocation="http://www.geni.net/resources/rspec/3 http://www.geni.net/resources/rspec/3/request.xsd " xmlns:client="http://www.protogeni.net/resources/rspec/ext/client/1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns="http://www.geni.net/resources/rspec/3">
  <node client_id="PC1" component_manager_id="urn:publicid:IDN+ilabt.imec.be+authority+cm" exclusive="false">
    <sliver_type name="container"/>
  </node>
  <node client_id="PC2" component_manager_id="urn:publicid:IDN+ilabt.imec.be+authority+cm" exclusive="false">
    <sliver_type name="container"/>
  </node>
</rspec>
`
	allocateArgs := &AllocateArgs{
		SliceURN:    testSliceIdentifier.URN(),
		Credentials: []Credential{testSliceCredential},
		Rspec:       rs,
	}
	allocateReply := &AllocateReply{}
	err := s.Allocate(r, allocateArgs, allocateReply)
	if err != nil {
		t.Errorf("Allocate() = %v; want nil", err)
	}
	got := allocateReply.Data.Code.Code
	if got != constants.GeniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, constants.GeniCodeSuccess)
	}
	deleteArgs := &DeleteArgs{
		URNs:        []string{testSliceIdentifier.URN()},
		Credentials: []Credential{testSliceCredential},
	}
	deleteReply := &DeleteReply{}
	err = s.Delete(r, deleteArgs, deleteReply)
	if err != nil {
		t.Errorf("Allocate() = %v; want nil", err)
	}
	got = deleteReply.Data.Code.Code
	if got != constants.GeniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, constants.GeniCodeSuccess)
	}
	// Verify deployments
	deployments, err := s.Deployments().List(context.TODO(), v1.ListOptions{})
	utils.Check(err)
	if len(deployments.Items) != 0 {
		t.Errorf("len(deployments) = %d; want 0", len(deployments.Items))
	}
}
