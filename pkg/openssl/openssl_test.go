package openssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	"math/big"
	"testing"
	"time"
)

func TestVerify(t *testing.T) {
	// TODO: Cleanup this test.
	// TODO: Use certificates signed/non-signed by root for proper testing.
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
	err = Verify([][]byte{derBytes}, derBytes)
	if err != nil {
		t.Errorf("Verify() = %s; want nil", err)
	}
	err = Verify([][]byte{}, derBytes)
	if err == nil {
		t.Errorf("Verify() = nil; want error")
	}
}
