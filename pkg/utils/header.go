package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"strings"
)

const (
	HttpHeaderCertificate = "X-Fed4Fire-Certificate"
	HttpHeaderUser        = "X-Fed4Fire-User"
)

func GetUserUrn(pemEncodedCert []byte) (string, error) {
	block, _ := pem.Decode([]byte(pemEncodedCert))
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
