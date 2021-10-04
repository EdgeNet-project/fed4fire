package service

import (
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	"github.com/EdgeNet-project/fed4fire/pkg/sfa"
	"github.com/EdgeNet-project/fed4fire/pkg/urn"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	"html"
	"io"
	"k8s.io/client-go/kubernetes"
	"log"
	"os/exec"
)

type Service struct {
	AbsoluteURL          string
	AuthorityName        string
	ContainerImages      map[string]string
	ContainerCpuLimit    string
	ContainerMemoryLimit string
	NamespaceCpuLimit    string
	NamespaceMemoryLimit string
	ParentNamespace      string
	EdgenetClient        *versioned.Clientset
	KubernetesClient     *kubernetes.Clientset
}

func (s Service) URN(resourceType string, resourceName string) string {
	identifier := urn.Identifier{
		Authorities:  []string{s.AuthorityName},
		ResourceType: resourceType,
		ResourceName: resourceName,
	}
	return identifier.String()
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

//func XmlSec1Verify
//To validate a credential:
//
//Credentials must validate against the credential schema.
//The credential signature must be valid, as per the ​XML Digital Signature standard.
//All contained certificates must be valid and trusted (trace back through all valid/trusted certificates to a trusted root certificate), and follow the GENI Certificate restrictions (see GeniApiCertificates).
//The expiration of the credential and all contained certificates must be later than the current time.
//All contained URNs must follow the GENI URN rules.
//The same rules apply to any parent credential, if the credential is delegated (and on up the delegation chain).
//For non delegated credentials, or for the root credential of a delegated credential (all the way back up any delegation chain), the signer must have authority over the target. Specifically, the credential issuer must have a URN indicating it is of type authority, and it must be the toplevelauthority or a parent authority of the authority named in the credential's target URN. See the URN rules page for details about authorities.
//For delegated credentials, the signer of the credential must be the subject (owner) of the parent credential), until you get to the root credential (no parent), in which case the above rule applies.

func (c Credential) Validate() bool {
	// TODO: Accept path to PEMs
	cmd := exec.Command("xmlsec1", "--verify",
		"--trusted-pem", "/Users/maxmouchet/Clones/github.com/EdgeNet-project/fed4fire/trusted_roots/ilabt.imec.be.pem",
		"-")
	stdin, err := cmd.StdinPipe()
	utils.Check(err)

	io.WriteString(stdin, html.UnescapeString(c.Value))
	stdin.Close()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", out)
	return true
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
	defaultCpuRequest    = "0.01"
	defaultMemoryRequest = "16Mi"
	defaultPauseImage    = "k8s.gcr.io/pause:latest"
)

const (
	fed4fireClientId   = "fed4fire.eu/client-id"
	fed4fireExpiryTime = "fed4fire.eu/expiry-time"
	fed4fireImageName  = "fed4fire.eu/image-name"
	fed4fireSlice      = "fed4fire.eu/slice"
	fed4fireUser       = "fed4fire.eu/user"
)

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

const (
	geniAllocateMany = "geny_many"
)

const (
	geniCredentialTypeSfa = "geny_sfa"
)
