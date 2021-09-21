package service

import (
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/sfa"
	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
	"html"
	"k8s.io/client-go/kubernetes"
)

type Service struct {
	AbsoluteURL      string
	AuthorityName    string
	KubernetesClient *kubernetes.Clientset
}

func (s Service) URN(type_ string, name string) string {
	// https://groups.geni.net/geni/wiki/GeniApiIdentifiers
	// The format of a GENI URN is: urn:publicid:IDN+<authority string>+<type>+<name>
	return fmt.Sprintf("urn:publicid:IDN+%s+%s+%s", s.AuthorityName, type_, name)
}

type Code struct {
	Code int `xml:"geni_code"`
}

type Credential struct {
	Type    string `xml:"geni_type"`
	Version string `xml:"geni_version"`
	Value   string `xml:"geni_value"`
}

func (c Credential) SFA() sfa.SignedCredential {
	if c.Type != "geni_sfa" {
		panic("credential type is not geni_sfa")
	}
	v := sfa.SignedCredential{}
	err := xml.Unmarshal([]byte(html.UnescapeString(c.Value)), &v)
	if err != nil {
		panic(err)
	}
	return v
}

func (c Credential) Validate() bool {
	//xml_ := html.UnescapeString(c.Value)
	//validator, err := signedxml.NewValidator(xml_)
	//fmt.Println(err)
	//xml1, err := validator.ValidateReferences()
	//fmt.Println(xml1)
	//fmt.Println(err)
	doc := etree.NewDocument()
	err := doc.ReadFromString(html.UnescapeString(c.Value))
	if err != nil {
		panic(err)
	}
	ctx := dsig.NewDefaultValidationContext(&dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	})
	ctx.IdAttribute = "xml:id"
	//TODO: Only use the returned validated element.
	//a := doc.ChildElements()[0].ChildElements()[0]
	a := doc.Root().ChildElements()[0]
	fmt.Println(a)
	_, err = ctx.Validate(a)
	fmt.Println(err)
	return false
}

type Options struct {
	Available    bool `xml:"geni_available"`
	Compressed   bool `xml:"geni_compressed"`
	RspecVersion struct {
		Type    string `xml:"type"`
		Version string `xml:"version"`
	} `xml:"geni_rspec_version"`
}

const (
	geniCodeSuccess               = 0
	geniCodeBadargs               = 1
	geniCodeError                 = 2
	geniCodeForbidden             = 3
	geniCodeBadversion            = 4
	geniCodeServerror             = 5
	geniCodeToobig                = 6
	geniCodeRefused               = 7
	geniCodeTimedout              = 8
	geniCodeDberror               = 9
	geniCodeRpcerror              = 10
	geniCodeUnavailable           = 11
	geniCodeSearchfailed          = 12
	geniCodeUnsupported           = 13
	geniCodeBusy                  = 14
	geniCodeExpired               = 15
	geniCodeInprogress            = 16
	geniCodeAlreadyexists         = 17
	geniCodeVlanUnavailable       = 24
	geniCodeInsufficientBandwidth = 25
)
