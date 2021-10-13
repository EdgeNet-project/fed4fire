package xmlsec1

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"
)

func TestSignVerify(t *testing.T) {
	// TODO: Cleanup this test.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	utils.Check(err)
	template := x509.Certificate{
		SerialNumber:          big.NewInt(0),
		Subject:               pkix.Name{CommonName: "example.org"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		privateKey.Public(),
		privateKey,
	)
	utils.Check(err)
	doc := `
<A>
<B xml:id="ref0"/>
<Signature xml:id="Sig_ref0" xmlns="http://www.w3.org/2000/09/xmldsig#">
	   <SignedInfo>
	     <CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"/>
	     <SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"/>
	     <Reference URI="#ref0">
	     <Transforms>
	       <Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature" />
	     </Transforms>
	     <DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/>
	     <DigestValue></DigestValue>
	     </Reference>
	   </SignedInfo>
	   <SignatureValue />
	     <KeyInfo>
	       <X509Data>
	         <X509SubjectName/>
	         <X509IssuerSerial/>
	         <X509Certificate/>
	       </X509Data>
	     <KeyValue />
	     </KeyInfo>
	   </Signature>
</A>
`
	res, err := Sign(*privateKey, derBytes, []byte(doc))
	if err != nil {
		t.Errorf("Sign() = %s; want nil", err)
	}
	err = Verify([][]byte{derBytes}, res)
	if err != nil {
		t.Errorf("Verify() = %s; want nil", err)
	}
	res = bytes.ReplaceAll(res, []byte("<B "), []byte("<C "))
	err = Verify([][]byte{derBytes}, res)
	if err == nil {
		t.Errorf("Verify() = nil; want error")
	}
}
