package service

import (
	"context"
	"encoding/pem"
	"encoding/xml"
	v1 "github.com/EdgeNet-project/fed4fire/pkg/apis/fed4fire/v1"
	"github.com/EdgeNet-project/fed4fire/pkg/generated/clientset/versioned"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	appsv1 "k8s.io/api/apps/v1"
	"net/http"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"

	"github.com/EdgeNet-project/fed4fire/pkg/sfa"
	"github.com/EdgeNet-project/fed4fire/pkg/xmlsec1"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	f4ftestclient "github.com/EdgeNet-project/fed4fire/pkg/generated/clientset/versioned/fake"
	"k8s.io/client-go/kubernetes"
	kubetestclient "k8s.io/client-go/kubernetes/fake"
)

var testAuthorityIdentifier = identifiers.MustParse("urn:publicid:IDN+example.org+authority+am")
var testAuthorityCaIdentifier = identifiers.MustParse("urn:publicid:IDN+example.org+authority+ca")
var testSliceIdentifier = identifiers.MustParse("urn:publicid:IDN+example.org+slice+test")
var testUserIdentifier = identifiers.MustParse("urn:publicid:IDN+example.org+user+test")

var authorityCert, authorityKey = utils.CreateCertificate(
	"authority.localhost",
	"authority@localhost",
	testAuthorityCaIdentifier.URN(),
	nil,
	nil,
)

var userCert, _ = utils.CreateCertificate(
	"test",
	"test@localhost",
	testUserIdentifier.URN(),
	authorityCert,
	authorityKey,
)

var testSliceCredential = createCredential(testUserIdentifier, testSliceIdentifier)

const testRspecSingle = `<rspec type="request" generated="2013-01-16T14:20:39Z" xsi:schemaLocation="http://www.geni.net/resources/rspec/3 http://www.geni.net/resources/rspec/3/request.xsd " xmlns:client="http://www.protogeni.net/resources/rspec/ext/client/1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns="http://www.geni.net/resources/rspec/3">
  <node client_id="PC1" component_manager_id="urn:publicid:IDN+example.org+authority+am" exclusive="false">
    <sliver_type name="container"/>
  </node>
</rspec>`

const testRspecMany = `<rspec type="request" generated="2013-01-16T14:20:39Z" xsi:schemaLocation="http://www.geni.net/resources/rspec/3 http://www.geni.net/resources/rspec/3/request.xsd " xmlns:client="http://www.protogeni.net/resources/rspec/ext/client/1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns="http://www.geni.net/resources/rspec/3">
  <node client_id="PC1" component_manager_id="urn:publicid:IDN+example.org+authority+am" exclusive="false">
    <sliver_type name="container"/>
  </node>
  <node client_id="PC2" component_id="urn:publicid:IDN+example.org+node+node-1" component_manager_id="urn:publicid:IDN+example.org+authority+am" exclusive="false">
    <sliver_type name="container"/>
    <hardware_type>
      <name>amd64</name>
	</hardware_type>
  </node>
</rspec>
`

func testService() *Service {
	var fed4fireClient versioned.Interface = f4ftestclient.NewSimpleClientset()
	var kubernetesClient kubernetes.Interface = kubetestclient.NewSimpleClientset()
	return &Service{
		AuthorityIdentifier: testAuthorityIdentifier,
		ContainerImages: map[string]string{
			"ubuntu2004": "docker.io/library/ubuntu:20.04",
		},
		ContainerCpuLimit:    "2",
		ContainerMemoryLimit: "2Gi",
		NamespaceCpuLimit:    "8",
		NamespaceMemoryLimit: "8Gi",
		Fed4FireClient:       fed4fireClient,
		KubernetesClient:     kubernetesClient,
		TrustedCertificates:  [][]byte{authorityCert},
	}
}

func testRequest() *http.Request {
	r, _ := http.NewRequestWithContext(context.TODO(), "", "", nil)
	r.Header.Set(constants.HttpHeaderUser, testUserIdentifier.URN())
	return r
}

func testNode(name string, ready bool) *corev1.Node {
	var readyStatus corev1.ConditionStatus = "True"
	if !ready {
		readyStatus = "False"
	}
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"edge-net.io/lat": "s-23.533500",
				"edge-net.io/lon": "w-46.635900",
			},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   "Ready",
					Status: readyStatus,
				},
			},
		},
	}
}

func allocateTestSlice(service *Service, request *http.Request, rspec string) {
	args := &AllocateArgs{
		SliceURN:    testSliceIdentifier.URN(),
		Credentials: []Credential{testSliceCredential},
		Rspec:       rspec,
	}
	reply := &AllocateReply{}
	err := service.Allocate(request, args, reply)
	utils.Check(err)
}

func provisionTestSlice(service *Service, request *http.Request) {
	args := &ProvisionArgs{
		URNs:        []string{testSliceIdentifier.URN()},
		Credentials: []Credential{testSliceCredential},
	}
	reply := &ProvisionReply{}
	err := service.Provision(request, args, reply)
	utils.Check(err)
}

func listTestDeployments(service *Service) []appsv1.Deployment {
	deployments, err := service.Deployments().List(context.TODO(), metav1.ListOptions{})
	utils.Check(err)
	return deployments.Items
}

func listTestSlivers(service *Service) []v1.Sliver {
	slivers, err := service.Slivers().List(context.TODO(), metav1.ListOptions{})
	utils.Check(err)
	return slivers.Items
}

func unmarshalTestRspec(s string) rspec.Rspec {
	v := rspec.Rspec{}
	err := xml.Unmarshal([]byte(s), &v)
	if err != nil {
		err := xml.Unmarshal(utils.DecompressZlibBase64(s), &v)
		utils.Check(err)
	}
	return v
}

func createCredential(owner identifiers.Identifier, target identifiers.Identifier) Credential {
	ownerCert, _ := utils.CreateCertificate(
		owner.URN(),
		"",
		owner.URN(),
		authorityCert,
		authorityKey,
	)
	targetCert, _ := utils.CreateCertificate(
		target.URN(),
		"",
		target.URN(),
		authorityCert,
		authorityKey,
	)
	ownerGid := pem.EncodeToMemory(&pem.Block{
		Type:  utils.PEMBlockTypeCertificate,
		Bytes: ownerCert,
	})
	targetGid := pem.EncodeToMemory(&pem.Block{
		Type:  utils.PEMBlockTypeCertificate,
		Bytes: targetCert,
	})
	credential := sfa.Credential{
		Id:        "ref0",
		Type:      "privilege",
		Serial:    "1",
		OwnerGID:  string(ownerGid),
		OwnerURN:  owner.URN(),
		TargetGID: string(targetGid),
		TargetURN: target.URN(),
		Expires:   time.Now().Add(1 * time.Hour),
	}
	unsignedCredential := sfa.SignedCredential{
		Credential: credential,
		Signatures: []sfa.Signature{{InnerXML: xmlsec1.Template}},
	}
	unsignedCredentialBytes, err := xml.Marshal(unsignedCredential)
	if err != nil {
		panic(err)
	}
	signedCredentialBytes, err := xmlsec1.Sign(
		*authorityKey,
		authorityCert,
		unsignedCredentialBytes,
	)
	if err != nil {
		panic(err)
	}
	return Credential{
		Type:    constants.GeniCredentialTypeSfa,
		Version: "3",
		Value:   string(signedCredentialBytes),
	}
}
