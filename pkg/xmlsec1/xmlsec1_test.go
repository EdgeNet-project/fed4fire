package xmlsec1

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"
)

func TestSignVerify(t *testing.T) {
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
	doc := fmt.Sprintf("<A><B xml:id=\"ref0\"/>%s</A>", Template)
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
