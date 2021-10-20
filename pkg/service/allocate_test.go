package service

import (
	"context"
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	if got != geniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, geniCodeSuccess)
	}
	deployments, err := s.KubernetesClient.AppsV1().
		Deployments(s.Namespace).
		List(context.TODO(), v1.ListOptions{})
	utils.Check(err)
	if len(deployments.Items) != 1 {
		t.Errorf("len(deployments) = %d; want 1", len(deployments.Items))
	}
	servicesClient := s.KubernetesClient.CoreV1().Services(s.Namespace)
	services, err := servicesClient.List(context.TODO(), v1.ListOptions{})
	utils.Check(err)
	if len(services.Items) != 1 {
		t.Errorf("len(services) = %d; want 1", len(services.Items))
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
	if got != geniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, geniCodeSuccess)
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
	if got != geniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, geniCodeSuccess)
	}
	// Verify deployments
	deploymentsClient := s.KubernetesClient.AppsV1().Deployments(s.Namespace)
	deployments, err := deploymentsClient.List(context.TODO(), v1.ListOptions{})
	utils.Check(err)
	if len(deployments.Items) != 2 {
		t.Errorf("len(deployments) = %d; want 2", len(deployments.Items))
	}
	servicesClient := s.KubernetesClient.CoreV1().Services(s.Namespace)
	services, err := servicesClient.List(context.TODO(), v1.ListOptions{})
	utils.Check(err)
	if len(services.Items) != 2 {
		t.Errorf("len(services) = %d; want 2", len(services.Items))
	}
}
