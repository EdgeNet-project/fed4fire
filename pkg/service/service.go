package service

import "k8s.io/client-go/kubernetes"

type Service struct {
	AbsoluteURL      string
	URN              string
	KubernetesClient *kubernetes.Clientset
}

type Code struct {
	Code int `xml:"geni_code"`
}

type Credential struct {
	Type    string `xml:"geni_type"`
	Version int    `xml:"geni_version"`
	Value   string `xml:"geni_value"`
}

type Options struct {
	Available    bool `xml:"geni_available"`
	Compressed   bool `xml:"geni_compressed"`
	RspecVersion struct {
		Type    string
		Version string
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
