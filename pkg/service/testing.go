package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	edgenettestclient "github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned/fake"
	"k8s.io/client-go/kubernetes"
	kubetestclient "k8s.io/client-go/kubernetes/fake"
)

const testAuthorityCaUrn = "urn:publicid:IDN+example.org+authority+ca"
const testSliceUrn = "urn:publicid:IDN+example.org+slice+test"
const testUserUrn = "urn:publicid:IDN+example.org+user+test"

func testService() *Service {
	var edgenetClient versioned.Interface = edgenettestclient.NewSimpleClientset()
	var kubernetesClient kubernetes.Interface = kubetestclient.NewSimpleClientset()
	return &Service{
		ContainerImages: map[string]string{
			"ubuntu2004": "docker.io/library/ubuntu:20.04",
		},
		EdgenetClient:    edgenetClient,
		KubernetesClient: kubernetesClient,
	}
}

func testRequest() *http.Request {
	r, _ := http.NewRequestWithContext(context.TODO(), "", "", nil)
	r.Header.Set(utils.HttpHeaderUser, testUserUrn)
	return r
}

func testNode(name string, ready bool) *v1.Node {
	var readyStatus v1.ConditionStatus = "True"
	if !ready {
		readyStatus = "False"
	}
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"edge-net.io/lat": "s-23.533500",
				"edge-net.io/lon": "w-46.635900",
			},
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   "Ready",
					Status: readyStatus,
				},
			},
		},
	}
}

func randSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	utils.Check(err)
	return serialNumber
}

func parseUrl(s string) *url.URL {
	url_, err := url.Parse(s)
	utils.Check(err)
	return url_
}

// TODO: Refactor and move all of this to a `testutils` package?
type certificateAuthority struct {
	rootCertificate []byte
	privateKey      []byte
}

func generateCertificateAuthority(fqdn string) certificateAuthority {
	urn := parseUrl(fmt.Sprintf("urn:publicid:IDN+%s+authority+ca", fqdn))
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	utils.Check(err)
	template := x509.Certificate{
		SerialNumber:          randSerialNumber(),
		Subject:               pkix.Name{CommonName: fqdn},
		URIs:                  []*url.URL{urn},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		publicKey,
		privateKey,
	)
	utils.Check(err)
	return certificateAuthority{
		rootCertificate: derBytes,
		privateKey:      privateKey,
	}
}

//func (c certificateAuthority) generateCertificate() []byte

func testCredential(
	authorityKey rsa.PrivateKey,
	authorityCertificate []byte,
	ownerUrn string,
	targetUrn string,
) {
	// TODO: Generate ownerGid and targetGid
	// TODO: Concatenate SignedCredential to template?
	//	template := `
	//<Signature xml:id="Sig_ref0" xmlns="http://www.w3.org/2000/09/xmldsig#">
	//	<SignedInfo>
	//		<CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"/>
	//		<SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"/>
	//		<Reference URI="#ref0">
	//			<Transforms>
	//				<Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature" />
	//			</Transforms>
	//			<DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/>
	//			<DigestValue></DigestValue>
	//		</Reference>
	//	</SignedInfo>
	//	<SignatureValue/>
	//	<KeyInfo>
	//		<X509Data>
	// 			<X509SubjectName/>
	// 			<X509IssuerSerial/>
	// 			<X509Certificate/>
	//		</X509Data>
	//		<KeyValue/>
	//	</KeyInfo>
	//</Signature>
	//`
	//func MarshalPKCS1PrivateKey(key *rsa.PrivateKey) []byte
	//xmlsec.Sign(authorityKey.)
	//type Credential struct {
	//	XMLName    xml.Name   `xml:"credential"`
	//	Type       string     `xml:"type"`
	//	Serial     string     `xml:"serial"`
	//	OwnerGID   string     `xml:"owner_gid"`
	//	OwnerURN   string     `xml:"owner_urn"`
	//	TargetGID  string     `xml:"target_gid"`
	//	TargetURN  string     `xml:"target_urn"`
	//	Expires    time.Time  `xml:"expires"`
	//	Privileges Privileges `xml:"privileges"`
	//}
}
