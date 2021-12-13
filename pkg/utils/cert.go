package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/url"
	"time"
)

func CreateCertificate(
	commonName string,
	emailAddress string,
	identifier string,
	parentCertificate []byte,
	parentPrivateKey *rsa.PrivateKey,
) ([]byte, *rsa.PrivateKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	serialNumber, err := randSerialNumber()
	if err != nil {
		panic(err)
	}
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	if emailAddress != "" {
		template.EmailAddresses = []string{emailAddress}
	}
	if identifier != "" {
		identifierUrl, err := url.Parse(identifier)
		if err != nil {
			panic(err)
		}
		template.URIs = []*url.URL{identifierUrl}
	}
	var parentCertificate_ *x509.Certificate
	var parentPrivateKey_ *rsa.PrivateKey
	if parentCertificate == nil {
		parentCertificate_ = template
		parentPrivateKey_ = privateKey
		template.KeyUsage |= x509.KeyUsageCertSign
		template.IsCA = true
	} else {
		parentCertificate_, err = x509.ParseCertificate(parentCertificate)
		if err != nil {
			panic(err)
		}
		parentPrivateKey_ = parentPrivateKey
	}
	der, err := x509.CreateCertificate(
		rand.Reader,
		template,
		parentCertificate_,
		privateKey.Public(),
		parentPrivateKey_,
	)
	if err != nil {
		panic(err)
	}
	return der, privateKey
}

func randSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	return serialNumber, err
}
