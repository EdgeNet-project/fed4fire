package sfa

import (
	"encoding/xml"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
)

type SignedCredential struct {
	XMLName    xml.Name   `xml:"signed-credential"`
	Credential Credential `xml:"credential"`
	// TODO: Add Signature infos
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

//type Signature struct {
//	XMLName xml.Name `xml:"Signature"`
//	Id      string   `xml:"id,attr"`
//}
//
//type SignedInfo struct {
//	XMLName                xml.Name               `xml:"SignedInfo"`
//	CanonicalizationMethod CanonicalizationMethod `xml:"CanonicalizationMethod"`
//}
//
//type CanonicalizationMethod struct {
//	XMLName   xml.Name `xml:"CanonicalizationMethod"`
//	Algorithm string   `xml:"Algorithm,attr"`
//}
//
//type SignatureMethod struct {
//	XMLName   xml.Name `xml:"SignatureMethod"`
//	Algorithm string   `xml:"Algorithm,attr"`
//}

func (c Credential) OwnerIdentifier() (*identifiers.Identifier, error) {
	return identifiers.Parse(c.OwnerURN)
}

func (c Credential) TargetIdentifier() (*identifiers.Identifier, error) {
	return identifiers.Parse(c.TargetURN)
}

//func (c Credential) OwnerCertificate() *x509.Certificate {
//	block, _ := pem.Decode([]byte(c.OwnerGID))
//	if block == nil {
//		panic("failed to parse certificate PEM")
//	}
//	cert, err := x509.ParseCertificate(block.Bytes)
//	if err != nil {
//		panic(err)
//	}
//	return cert
//}
//
//func (c Credential) TargetCertificate() *x509.Certificate {
//	block, _ := pem.Decode([]byte(c.TargetGID))
//	if block == nil {
//		panic("failed to parse certificate PEM")
//	}
//	cert, err := x509.ParseCertificate(block.Bytes)
//	if err != nil {
//		panic(err)
//	}
//	return cert
//}
