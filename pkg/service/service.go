package service

import (
	"fmt"
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
