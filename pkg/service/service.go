// Package service implements the XML-RPC methods specified by the AM API.
package service

import (
	"context"
	"encoding/xml"
	"fmt"
	v1 "github.com/EdgeNet-project/fed4fire/pkg/apis/fed4fire/v1"
	"github.com/EdgeNet-project/fed4fire/pkg/generated/clientset/versioned"
	fed4firev1 "github.com/EdgeNet-project/fed4fire/pkg/generated/clientset/versioned/typed/fed4fire/v1"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	"html"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"net/http"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/EdgeNet-project/fed4fire/pkg/naming"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/EdgeNet-project/fed4fire/pkg/openssl"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/EdgeNet-project/fed4fire/pkg/xmlsec1"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	"github.com/EdgeNet-project/fed4fire/pkg/sfa"
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
	Namespace            string
	TrustedCertificates  [][]byte
	Fed4FireClient       versioned.Interface
	KubernetesClient     kubernetes.Interface
}

func (s Service) ConfigMaps() typedcorev1.ConfigMapInterface {
	return s.KubernetesClient.CoreV1().ConfigMaps(s.Namespace)
}

func (s Service) Deployments() typedappsv1.DeploymentInterface {
	return s.KubernetesClient.AppsV1().Deployments(s.Namespace)
}

func (s Service) Nodes() typedcorev1.NodeInterface {
	return s.KubernetesClient.CoreV1().Nodes()
}

func (s Service) Pods() typedcorev1.PodInterface {
	return s.KubernetesClient.CoreV1().Pods(s.Namespace)
}

func (s Service) Services() typedcorev1.ServiceInterface {
	return s.KubernetesClient.CoreV1().Services(s.Namespace)
}

func (s Service) Slivers() fed4firev1.SliverInterface {
	return s.Fed4FireClient.Fed4fireV1().Slivers(s.Namespace)
}

func (s Service) GetSliver(ctx context.Context, name string) *v1.Sliver {
	sliver, err := s.Slivers().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.ErrorS(err, "Failed to get sliver")
		}
		return nil
	}
	return sliver
}

func (s Service) GetSliverArchHostPort(ctx context.Context, name string) (*string, *string, *int) {
	service, err := s.Services().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.ErrorS(err, "Failed to get service")
		}
		return nil, nil, nil
	}
	pods, err := s.Pods().List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("%s=%s", "status.phase", corev1.PodRunning),
		LabelSelector: fmt.Sprintf("%s=%s", constants.Fed4FireSliverName, name),
	})
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.ErrorS(err, "Failed to list pods")
		}
		return nil, nil, nil
	}
	if len(pods.Items) > 0 {
		pod := pods.Items[0]
		node, err := s.Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				klog.ErrorS(err, "Failed to get node")
			}
			return nil, nil, nil
		}
		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeInternalIP {
				return pointer.StringPtr(
						node.Labels[corev1.LabelArchStable],
					), pointer.StringPtr(
						address.Address,
					), pointer.IntPtr(
						int(service.Spec.Ports[0].NodePort),
					)
			}
		}
	}
	return nil, nil, nil
}

func (s Service) GetSliverDeployment(ctx context.Context, name string) *appsv1.Deployment {
	deployment, err := s.Deployments().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.ErrorS(err, "Failed to get deployment")
		}
		return nil
	}
	return deployment
}

func (s Service) GetSliverStatus(ctx context.Context, name string) (string, string) {
	allocationStatus := constants.GeniStateUnallocated
	operationalStatus := constants.GeniStateNotReady
	sliver := s.GetSliver(ctx, name)
	if sliver != nil {
		allocationStatus = constants.GeniStateAllocated
		deployment := s.GetSliverDeployment(ctx, name)
		if deployment != nil {
			allocationStatus = constants.GeniStateProvisioned
		}
		arch, host, port := s.GetSliverArchHostPort(ctx, name)
		if arch != nil && host != nil && port != nil {
			operationalStatus = constants.GeniStateReady
		}
	}
	return allocationStatus, operationalStatus
}

func (s Service) ListSlivers(
	ctx context.Context,
	identifier identifiers.Identifier,
) ([]v1.Sliver, error) {
	switch identifier.ResourceType {
	case identifiers.ResourceTypeSlice:
		sliceHash := naming.SliceHash(identifier.URN())
		labelSelector := fmt.Sprintf("%s=%s", constants.Fed4FireSliceHash, sliceHash)
		slivers, err := s.Slivers().List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return nil, err
		}
		return slivers.Items, nil
	case identifiers.ResourceTypeSliver:
		sliver, err := s.Slivers().Get(ctx, identifier.ResourceName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return []v1.Sliver{*sliver}, nil
	default:
		return nil, fmt.Errorf("identifier type must be slice or sliver")
	}
}

func (s Service) AuthorizeAndListSlivers(
	r *http.Request,
	resourceIdentifiersStr []string,
	credentials []Credential,
) ([]v1.Sliver, error) {
	userIdentifierStr := r.Header.Get(constants.HttpHeaderUser)
	userIdentifier, err := identifiers.Parse(userIdentifierStr)
	if err != nil {
		return nil, err
	}
	slivers := make([]v1.Sliver, 0)
	for _, urn := range resourceIdentifiersStr {
		identifier, err := identifiers.Parse(urn)
		if err != nil {
			return nil, err
		}
		// Verify that a user is authorized for a slice before listing the slivers inside.
		// This is mostly to be compatible with the spec. that expects an error on an
		// un-existing slice (instead of a list of 0 slivers).
		if identifier.ResourceType == identifiers.ResourceTypeSlice {
			_, err := FindCredential(
				*userIdentifier,
				identifier,
				credentials,
				s.TrustedCertificates,
			)
			if err != nil {
				return nil, err
			}
		}
		slivers_, err := s.ListSlivers(r.Context(), *identifier)
		if err != nil {
			return nil, err
		}
		slivers = append(slivers, slivers_...)
	}
	for _, sliver := range slivers {
		_, err := FindCredentialForSliver(
			*userIdentifier,
			sliver,
			credentials,
			s.TrustedCertificates,
		)
		if err != nil {
			return nil, err
		}
	}
	return slivers, nil
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

type RspecVersion struct {
	Type       string   `xml:"type"`
	Version    string   `xml:"version"`
	Schema     string   `xml:"schema"`
	Namespace  string   `xml:"namespace"`
	Extensions []string `xml:"extensions"`
}

type Sliver struct {
	URN               string `xml:"geni_sliver_urn"`
	Expires           string `xml:"geni_expires"`
	AllocationStatus  string `xml:"geni_allocation_status"`
	OperationalStatus string `xml:"geni_operational_status"`
	Error             string `xml:"geni_error"`
}

type Options struct {
	// XML-RPC boolean value indicating whether the caller is interested in
	// all resources or available resources.
	// If this value is true (1), the result should contain only available resources.
	// If this value is false (0) or unspecified, both available and allocated resources should be returned.
	// The Aggregate Manager is free to limit visibility of certain resources based on the credentials parameter.
	Available  bool `xml:"geni_available"`
	BestEffort bool `xml:"geni_best_effort"`
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
	RspecVersion RspecVersion `xml:"geni_rspec_version"`
	Users        []struct {
		URN  string   `xml:"urn"`
		Keys []string `xml:"keys"`
	} `xml:"geni_users"`
}

func NewSliver(sliver v1.Sliver, allocationStatus string, operationalStatus string) Sliver {
	return Sliver{
		URN:               sliver.Spec.URN,
		Expires:           sliver.Spec.Expires.Format(time.RFC3339),
		AllocationStatus:  allocationStatus,
		OperationalStatus: operationalStatus,
	}
}

func (c Credential) ValidatedSFA(trustedCertificates [][]byte) (*sfa.Credential, error) {
	if c.Type != constants.GeniCredentialTypeSfa {
		return nil, fmt.Errorf("credential type is not geni_sfa")
	}
	val := []byte(html.UnescapeString(c.Value))
	// 1. Verify the credential signature
	err := xmlsec1.Verify(trustedCertificates, val)
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
	err = openssl.Verify(trustedCertificates, utils.PEMDecodeMany([]byte(v.Credential.OwnerGID)))
	if err != nil {
		return nil, err
	}
	err = openssl.Verify(trustedCertificates, utils.PEMDecodeMany([]byte(v.Credential.TargetGID)))
	if err != nil {
		return nil, err
	}
	// 4. Verify expiration time
	if v.Credential.Expires.Before(time.Now()) {
		return nil, fmt.Errorf("credential has expired")
	}
	// TODO: Handle delegation:
	// For non delegated credentials, or for the root credential of a delegated credential (all the way back up any delegation chain), the signer must have authority over the target. Specifically, the credential issuer must have a URN indicating it is of type authority, and it must be the toplevelauthority or a parent authority of the authority named in the credential's target URN. See the URN rules page for details about authorities.
	// For delegated credentials, the signer of the credential must be the subject (owner) of the parent credential), until you get to the root credential (no parent), in which case the above rule applies.
	return &v.Credential, nil
}

func FindCredential(
	userIdentifier identifiers.Identifier,
	targetIdentifier *identifiers.Identifier,
	credentials []Credential,
	trustedCertificates [][]byte,
) (*sfa.Credential, error) {
	for _, credential := range credentials {
		if credential.Type != constants.GeniCredentialTypeSfa {
			continue
		}
		validated, err := credential.ValidatedSFA(trustedCertificates)
		if err != nil {
			return nil, err
		}
		ownerId, err := identifiers.Parse(validated.OwnerURN)
		if err != nil {
			return nil, err
		}
		targetId, err := identifiers.Parse(validated.TargetURN)
		if err != nil {
			return nil, err
		}
		if targetIdentifier == nil && ownerId.Equal(userIdentifier) {
			return validated, nil
		} else if targetIdentifier != nil && ownerId.Equal(userIdentifier) && targetId.Equal(*targetIdentifier) {
			return validated, nil
		}
	}
	return nil, fmt.Errorf("no matching credential found")
}

func FindCredentialForSliver(
	userIdentifier identifiers.Identifier,
	sliver v1.Sliver,
	credentials []Credential,
	trustedCertificates [][]byte,
) (*sfa.Credential, error) {
	sliverIdentifier, err := identifiers.Parse(sliver.Spec.URN)
	if err != nil {
		return nil, err
	}
	sliceIdentifier, err := identifiers.Parse(sliver.Spec.SliceURN)
	if err != nil {
		return nil, err
	}
	credential, err := FindCredential(
		userIdentifier,
		sliverIdentifier,
		credentials,
		trustedCertificates,
	)
	if err == nil {
		return credential, nil
	}
	credential, err = FindCredential(
		userIdentifier,
		sliceIdentifier,
		credentials,
		trustedCertificates,
	)
	if err == nil {
		return credential, nil
	}
	return nil, fmt.Errorf("no matching credential found")
}

func MarshalRspec(rspec rspec.Rspec, compressed bool) (string, error) {
	b, err := xml.Marshal(rspec)
	if err != nil {
		return "", err
	}
	if compressed {
		return utils.CompressZlibBase64(b), nil
	}
	return string(b), nil
}
