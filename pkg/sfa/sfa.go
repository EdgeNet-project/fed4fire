package sfa

import (
	"crypto/x509"
	"encoding/pem"
	"encoding/xml"
	"time"
)

type SignedCredential struct {
	XMLName    xml.Name   `xml:"signed-credential"`
	Credential Credential `xml:"credential"`
}

type Credential struct {
	XMLName    xml.Name   `xml:"credential"`
	Type       string     `xml:"type"`
	Serial     string     `xml:"serial"`
	OwnerGID   string     `xml:"owner_gid"`
	OwnerURN   string     `xml:"owner_urn"`
	TargetGID  string     `xml:"target_gid"`
	TargetURN  string     `xml:"target_urn"`
	Expires    time.Time  `xml:"expires"`
	Privileges Privileges `xml:"privileges"`
}

type Privileges struct {
	XMLName   xml.Name    `xml:"privileges"`
	Privilege []Privilege `xml:"privilege"`
}

type Privilege struct {
	XMLName     xml.Name `xml:"privilege"`
	Name        string   `xml:"name"`
	CanDelegate bool     `xml:"can_delegate"`
}

//type

func (c Credential) OwnerCertificate() *x509.Certificate {
	block, _ := pem.Decode([]byte(c.OwnerGID))
	if block == nil {
		panic("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	return cert
}
