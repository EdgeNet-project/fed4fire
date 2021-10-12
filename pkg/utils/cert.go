package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

func GetUserUrnFromHeader(header http.Header) (string, error) {
	pemEncodedCert, err := url.QueryUnescape(header.Get("X-Fed4Fire-Certificate"))
	if err != nil {
		return "", nil
	}
	return GetUserUrn([]byte(pemEncodedCert))
}
