// Package service implements the XML-RPC methods specified by the AM API.
package service

import (
	"encoding/xml"
	"fmt"
	"html"

	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	"github.com/EdgeNet-project/fed4fire/pkg/sfa"
	"github.com/crewjam/go-xmlsec"
	"k8s.io/client-go/kubernetes"
)

type Service struct {
	AbsoluteURL          string
	AuthorityIdentifier  identifiers.Identifier
	ContainerImages      map[string]string
	ContainerCpuLimit    string
	ContainerMemoryLimit string
	NamespaceCpuLimit    string
	NamespaceMemoryLimit string
	ParentNamespace      string
	TrustedCertificates  [][]byte
	EdgenetClient        versioned.Interface
	KubernetesClient     kubernetes.Interface
}

type Code struct {
	// An integer supplying the GENI standard return code indicating the success or failure of this call.
	// Error codes are standardized and defined in
	// https://groups.geni.net/geni/attachment/wiki/GAPI_AM_API_V3/CommonConcepts/geni-error-codes.xml.
	// Codes may be negative. A success return is defined as geni_code of 0.
	Code int `xml:"geni_code"`
}

type Credential struct {
	Type    string `xml:"geni_type"`
	Version string `xml:"geni_version"`
	Value   string `xml:"geni_value"`
}

type Sliver struct {
	URN              string `xml:"geni_sliver_urn"`
	Expires          string `xml:"geni_expires"`
	AllocationStatus string `xml:"geni_allocation_status"`
	Error            string `xml:"geni_error"`
}

func (c Credential) ValidatedSFA(trustedCertificates [][]byte) (*sfa.Credential, error) {
	if c.Type != "geni_sfa" {
		return nil, fmt.Errorf("credential type is not geni_sfa")
	}
	val := []byte(html.UnescapeString(c.Value))
	// 1. Verify the credential signature
	err := xmlsec.VerifyTrusted(trustedCertificates, val, xmlsec.SignatureOptions{})
	if err != nil {
		return nil, err
	}
	// 2. Decode the credential
	v := sfa.SignedCredential{}
	err = xml.Unmarshal(val, &v)
	if err != nil {
		return nil, err
	}
	// 3. Verify the embedded certificates
	// TODO
	// TODO: Handle delegation?
	// For non delegated credentials, or for the root credential of a delegated credential (all the way back up any delegation chain), the signer must have authority over the target. Specifically, the credential issuer must have a URN indicating it is of type authority, and it must be the toplevelauthority or a parent authority of the authority named in the credential's target URN. See the URN rules page for details about authorities.
	// For delegated credentials, the signer of the credential must be the subject (owner) of the parent credential), until you get to the root credential (no parent), in which case the above rule applies.
	return &v.Credential, nil
}

type Options struct {
	// XML-RPC boolean value indicating whether the caller is interested in
	// all resources or available resources.
	// If this value is true (1), the result should contain only available resources.
	// If this value is false (0) or unspecified, both available and allocated resources should be returned.
	// The Aggregate Manager is free to limit visibility of certain resources based on the credentials parameter.
	Available bool `xml:"geni_available"`
	// XML-RPC boolean value indicating whether the caller would like the result to be compressed.
	// If the value is true (1), the returned resource list will be compressed according to RFC 1950.
	// If the value is false (0) or unspecified, the return will be text.
	Compressed bool `xml:"geni_compressed"`
	// Requested expiration of all new slivers, may be ignored by aggregates.
	EndTime string `xml:"geni_end_time"`
	// XML-RPC struct indicating the type and version of Advertisement RSpec to return.
	// The struct contains 2 members, type and version. type and version are case-insensitive strings,
	// matching those in geni_ad_rspec_versions as returned by GetVersion at this aggregate.
	// This option is required, and aggregates are expected to return a geni_code of 1 (BADARGS) if it is missing.
	// Aggregates should return a geni_code of 4 (BADVERSION) if the requested RSpec version
	// is not one advertised as supported in GetVersion.
	RspecVersion struct {
		Type    string `xml:"type"`
		Version string `xml:"version"`
	} `xml:"geni_rspec_version"`
}

func FindMatchingCredential(
	userIdentifier identifiers.Identifier,
	targetIdentifier identifiers.Identifier,
	credentials []Credential,
	trustedCertificates [][]byte,
) (*sfa.Credential, error) {
	for _, credential := range credentials {
		validated, err := credential.ValidatedSFA(trustedCertificates)
		if err != nil {
			return nil, err
		}
		ownerId, err := validated.OwnerIdentifier()
		if err != nil {
			return nil, err
		}
		targetId, err := validated.TargetIdentifier()
		if err != nil {
			return nil, err
		}
		if ownerId.Equal(userIdentifier) && targetId.Equal(targetIdentifier) {
			return validated, nil
		}
	}
	return nil, fmt.Errorf("no matching credential found")
}

// Default value for new deployments.
const (
	defaultCpuRequest    = "0.01"
	defaultMemoryRequest = "16Mi"
)

// Names for Kubernetes objects labels and annotations.
const (
	fed4fireClientId = "fed4fire.eu/client-id"
	fed4fireExpires  = "fed4fire.eu/expires"
	fed4fireSlice    = "fed4fire.eu/slice"
	fed4fireSliver   = "fed4fire.eu/sliver"
	fed4fireUser     = "fed4fire.eu/user"
)

// https://groups.geni.net/geni/attachment/wiki/GAPI_AM_API_V3/CommonConcepts/geni-error-codes.xml
const (
	// Success
	geniCodeSuccess = 0
	// Bad Arguments: malformed
	geniCodeBadargs = 1
	// Error (other)
	geniCodeError = 2
	// Operation Forbidden: eg supplied credentials do not provide sufficient privileges (on the given slice)
	geniCodeForbidden = 3
	// Bad Version (eg of RSpec)
	geniCodeBadversion = 4
	// Server Error
	geniCodeServerror = 5
	// Too Big (eg request RSpec)
	geniCodeToobig = 6
	// Operation Refused
	geniCodeRefused = 7
	// Operation Timed Out
	geniCodeTimedout = 8
	// Database Error
	geniCodeDberror = 9
	// RPC Error
	geniCodeRpcerror = 10
	// Unavailable (eg server in lockdown)
	geniCodeUnavailable = 11
	// Search Failed (eg for slice)
	geniCodeSearchfailed = 12
	// Operation Unsupported
	geniCodeUnsupported = 13
	// Busy (resource, slice, or server); try again later
	geniCodeBusy = 14
	// Expired (eg slice)
	geniCodeExpired = 15
	// In Progress
	geniCodeInprogress = 16
	// Already Exists (eg slice)
	geniCodeAlreadyexists = 17
)

// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3/CommonConcepts#SliverAllocationStates
const (
	// The sliver does not exist. This is the small black circle in typical state diagrams.
	geniStateUnallocated = "geni_unallocated"
	// The sliver exists, defines particular resources, and is in a slice.
	// The aggregate has not (if possible) done any time consuming or expensive work to instantiate the resources,
	// provision them, or make it difficult to revert the slice to the state prior to allocating this sliver.
	// This state is what the aggregate is offering the experimenter.
	geniStateAllocated = "geni_allocated"
	// The aggregate has started instantiating resources, and otherwise making changes to resources
	// and the slice to make the resources available to the experimenter.
	// At this point, operational states are valid to specify further when
	// the resources are available for experimenter use.
	geniStateProvisioned = "geni_provisioned"
)

const (
	// Performing multiple Allocates without a delete is an error condition because the aggregate
	// only supports a single sliver per slice or does not allow incrementally adding new slivers.
	geniAllocateSingle = "geni_single"
	// Additional calls to Allocate must be disjoint from slivers allocated with previous calls
	// (no references or dependencies on existing slivers).
	// The topologies must be disjoint in that there can be no connection or other reference
	// from one topology to the other.
	geniAllocateDisjoint = "geni_disjoint"
	// Multiple slivers can exist and be incrementally added, including those which connect or overlap in some way.
	geniAllocateMany = "geny_many"
)

// https://groups.geni.net/geni/wiki/GeniApiCredentials
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3/CommonConcepts#credentials
const (
	geniCredentialTypeAbac = "geni_abac"
	geniCredentialTypeSfa  = "geny_sfa"
)
