package sfa

import (
	"encoding/xml"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	"github.com/EdgeNet-project/fed4fire/pkg/openssl"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
)

type SignedCredential struct {
	XMLName    xml.Name    `xml:"signed-credential"`
	Credential Credential  `xml:"credential"`
	Signatures []Signature `xml:"signatures"`
}

type Credential struct {
	XMLName    xml.Name   `xml:"credential"`
	Id         string     `xml:"http://www.w3.org/XML/1998/namespace id,attr,omitempty"`
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

type Signature struct {
	InnerXML string `xml:",innerxml"`
}

func (c Credential) Expired() bool {
	return c.Expires.Before(time.Now())
}

func (c Credential) ValidateOwner(trustedCertificates [][]byte) error {
	_, err := identifiers.Parse(c.OwnerURN)
	if err != nil {
		return err
	}
	certificateChain := utils.PEMDecodeMany([]byte(c.OwnerGID))
	return openssl.Verify(trustedCertificates, certificateChain)
}

func (c Credential) ValidateTarget(trustedCertificates [][]byte) error {
	_, err := identifiers.Parse(c.TargetURN)
	if err != nil {
		return err
	}
	certificateChain := utils.PEMDecodeMany([]byte(c.TargetGID))
	return openssl.Verify(trustedCertificates, certificateChain)
}
