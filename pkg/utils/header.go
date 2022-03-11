package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"strings"
)

func GetUserUrn(pemEncodedCert []byte) (string, error) {
	block, _ := pem.Decode(pemEncodedCert)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	for _, uri := range cert.URIs {
		if strings.HasPrefix(uri.String(), "urn:publicid:") {
			return uri.String(), nil
		}
	}
	return "", fmt.Errorf("user URN not found")
}

func GetUserUrnFromEscapedCert(escapedCert string) (string, error) {
	pemEncodedCert, err := url.QueryUnescape(escapedCert)
	if err != nil {
		return "", nil
	}
	return GetUserUrn([]byte(pemEncodedCert))
}
