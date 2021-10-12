package service

import (
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/sfa"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/EdgeNet-project/edgenet/pkg/apis/core/v1alpha"
	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

type AllocateOptions struct {
	// Optional. Requested expiration of all new slivers, may be ignored by aggregates.
	EndTime string `xml:"geni_end_time"`
}

type AllocateArgs struct {
	SliceURN    string
	Credentials []Credential
	Rspec       string
	Options     AllocateOptions
}

type AllocateReply struct {
	Data struct {
		Code  Code `xml:"code"`
		Value struct {
			Rspec   string   `xml:"geni_rspec"`
			Slivers []Sliver `xml:"geni_slivers"`
			Error   string   `xml:"geni_error"`
		} `xml:"value"`
	}
}

func (v *AllocateReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = geniCodeError
	v.Data.Value.Error = fmt.Sprintf("%s: %s", msg, err)
}

// Allocate allocates resources as described in a request RSpec argument to a slice with the named URN.
// On success, one or more slivers are allocated, containing resources satisfying the request, and assigned to the given slice.
// This method returns a listing and description of the resources reserved for the slice by this operation, in the form of a manifest RSpec.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Allocate
func (s *Service) Allocate(r *http.Request, args *AllocateArgs, reply *AllocateReply) error {
	// TODO: Return identifier instead? And move to this package?
	userUrn, err := utils.GetUserUrnFromHeader(r.Header)
	if err != nil {
		reply.SetAndLogError(err, "Failed to decoded user URN")
	}
	var matchingCredential *sfa.Credential
	for _, credential := range args.Credentials {
		validated, err := credential.ValidatedSFA(s.TrustedCertificates)
		if err != nil {
			reply.SetAndLogError(err, "Failed to validate credential")
			return nil
		}
		if validated.OwnerURN == userUrn && validated.TargetURN == args.SliceURN {
			matchingCredential = validated
			break
		}
	}
	if matchingCredential == nil {
		reply.SetAndLogError(fmt.Errorf("no matching credentials for user and slice URN"), "Invalid credentials")
		return nil
	}

	// TODO: Move to credential package? TargetIdentifier(), ...
	// TODO: Use sliceUrn instead (in case of delegation?)
	sliceIdentifier, err := identifiers.Parse(matchingCredential.TargetURN)
	if err != nil {
		reply.SetAndLogError(err, "Failed to parse slice URN")
	}

	// TODO: Use userUrn instead (in case of delegation?)
	userIdentifier, err := identifiers.Parse(matchingCredential.OwnerURN)
	if err != nil {
		reply.SetAndLogError(err, "Failed to parse user URN")
		return nil
	}

	requestRspec := rspec.Rspec{}
	err = xml.Unmarshal([]byte(html.UnescapeString(args.Rspec)), &requestRspec)
	if err != nil {
		reply.SetAndLogError(err, "Failed to deserialize rspec")
		return nil
	}

	// Q: How does a user choose a node?
	// A: By specifying the component ID. We must check that the node truly exists beforehand.
	// Actually require the user to specify a node name.
	// TODO: Fill missing values in the Rspec.

	// 1. Get or create the subnamespace for the slice
	subnamespaceClient := s.EdgenetClient.CoreV1alpha().SubNamespaces(s.ParentNamespace)
	subnamespaceName, err := subnamespaceNameForSlice(*sliceIdentifier)
	if err != nil {
		reply.SetAndLogError(
			err,
			"Failed to build subnamespace name from slice URN",
			"urn",
			args.SliceURN,
		)
		return nil
	}
	// 1.a. Get the subnamespace if it exists
	subnamespace, err := subnamespaceClient.Get(r.Context(), subnamespaceName, metav1.GetOptions{})
	// 1.b. Create the subnamespace if it doesn't exists
	if err != nil {
		klog.InfoS("Could not find subnamespace", "name", subnamespaceName)
		subnamespace, err = subnamespaceForSlice(
			*sliceIdentifier,
			s.NamespaceCpuLimit,
			s.NamespaceMemoryLimit,
		)
		if err != nil {
			reply.SetAndLogError(err, "Failed to build subnamespace", "name", subnamespaceName)
			return nil
		}
		_, err = subnamespaceClient.Create(r.Context(), subnamespace, metav1.CreateOptions{})
		if err != nil {
			reply.SetAndLogError(err, "Failed to create subnamespace", "name", subnamespaceName)
			return nil
		}
		klog.InfoS("Created subnamespace", "name", subnamespaceName)
	}

	// 2. Build the deployment objects
	deployments := make([]appsv1.Deployment, len(requestRspec.Nodes))
	for i, node := range requestRspec.Nodes {
		deployment, err := deploymentForRspec(
			node,
			s.AuthorityIdentifier,
			*sliceIdentifier,
			*userIdentifier,
			s.ContainerImages,
			s.ContainerCpuLimit,
			s.ContainerMemoryLimit,
		)
		if err != nil {
			reply.SetAndLogError(err, "Failed to build deployment")
			return nil
		}
		deployments[i] = *deployment
	}

	// 3. Create the deployment objects and rollback them in case of failure
	targetNamespace := fmt.Sprintf("%s-%s", s.ParentNamespace, subnamespaceName)
	deploymentsClient := s.KubernetesClient.AppsV1().Deployments(targetNamespace)
	deployed := make([]bool, len(deployments))
	success := true
	for i, deployment := range deployments {
		_, err := deploymentsClient.Create(r.Context(), &deployment, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			klog.InfoS("Ignoring already existing deployment", "name", deployment.Name)
			deployed[i] = false
		} else if err != nil {
			reply.SetAndLogError(err, "Failed to create deployment", "name", deployment.Name)
			deployed[i] = false
			success = false
			break
		} else {
			deployed[i] = true
		}
	}
	if !success {
		for i, isDeployed := range deployed {
			if isDeployed {
				deployment := deployments[i]
				klog.InfoS("Rolling back deployment", "name", deployment)
				err = deploymentsClient.Delete(r.Context(), deployment.Name, metav1.DeleteOptions{})
				if err != nil {
					klog.InfoS("Failed to rollback deployment", "name", deployment.Name)
				}
			}
		}
		return nil
	}

	// TODO: Specify the correct value for geni_allocate
	// As described here, the geni_allocate return from GetVersion advertises when a client may legally call Allocate
	// (only once at a time per slice, whenever desired, or multiple times only if the requested resources do not interact).
	returnRspec := rspec.Rspec{Type: rspec.RspecTypeRequest}
	for i, isDeployed := range deployed {
		if isDeployed {
			sliver := Sliver{
				URN:              deployments[i].Annotations[fed4fireSliver],
				Expires:          deployments[i].Annotations[fed4fireExpires],
				AllocationStatus: geniStateAllocated,
			}
			reply.Data.Value.Slivers = append(reply.Data.Value.Slivers, sliver)
			returnRspec.Nodes = append(returnRspec.Nodes, requestRspec.Nodes[i])
		}
	}

	xml_, err := xml.Marshal(returnRspec)
	if err != nil {
		reply.SetAndLogError(err, "Failed to serialize response")
		return nil
	}
	reply.Data.Value.Rspec = string(xml_)
	reply.Data.Code.Code = geniCodeSuccess
	return nil
}

func subnamespaceNameForSlice(identifier identifiers.Identifier) (string, error) {
	if identifier.ResourceType != identifiers.ResourceTypeSlice {
		return "", fmt.Errorf("URN resource type must be `slice`")
	}
	s := fmt.Sprintf(
		"fed4fire-slice-%s-%s",
		strings.Join(identifier.Authorities, "-"),
		identifier.ResourceName,
	)
	s = strings.ReplaceAll(s, ".", "-")
	return s, nil
}

func subnamespaceForSlice(
	identifier identifiers.Identifier,
	cpuLimit string,
	memoryLimit string,
) (*v1alpha.SubNamespace, error) {
	name, err := subnamespaceNameForSlice(identifier)
	if err != nil {
		return nil, err
	}
	subnamespace := &v1alpha.SubNamespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				fed4fireSlice: identifier.URN(),
			},
		},
		Spec: v1alpha.SubNamespaceSpec{
			Resources: v1alpha.Resources{
				CPU:    cpuLimit,
				Memory: memoryLimit,
			},
			Inheritance: v1alpha.Inheritance{
				NetworkPolicy: true,
				RBAC:          true,
			},
		},
	}
	return subnamespace, nil
}

func deploymentImageForSliverType(
	sliverType rspec.SliverType,
	containerImages map[string]string,
) (string, error) {
	if len(sliverType.DiskImages) == 0 {
		return utils.Keys(containerImages)[0], nil
	}
	if len(sliverType.DiskImages) == 1 {
		identifier, err := identifiers.Parse(sliverType.DiskImages[0].Name)
		if err != nil {
			return "", err
		}
		if image, ok := containerImages[identifier.ResourceName]; ok {
			return image, nil
		} else {
			return "", fmt.Errorf("invalid image name")
		}
	}
	return "", fmt.Errorf("no more than one disk image can be specified")
}

func deploymentForRspec(
	node rspec.Node,
	authorityIdentifier identifiers.Identifier,
	sliceIdentifier identifiers.Identifier,
	userIdentifier identifiers.Identifier,
	containerImages map[string]string,
	cpuLimit string,
	memoryLimit string,
) (*appsv1.Deployment, error) {
	if node.Exclusive {
		return nil, fmt.Errorf("exclusive must be false")
	}
	if len(node.SliverTypes) != 1 {
		return nil, fmt.Errorf("exactly one sliver type must be specified")
	}
	sliverType := node.SliverTypes[0]
	if sliverType.Name != "container" {
		return nil, fmt.Errorf("invalid sliver type")
	}
	image, err := deploymentImageForSliverType(sliverType, containerImages)
	if err != nil {
		return nil, err
	}
	clientId := strings.ToLower(node.ClientID)
	sliverIdentifier := authorityIdentifier.Copy(identifiers.ResourceTypeSliver,
		fmt.Sprintf(
			"%s-%s-%s",
			strings.Join(sliceIdentifier.Authorities, "-"),
			sliceIdentifier.ResourceName,
			clientId,
		),
	)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: clientId,
			Annotations: map[string]string{
				// TODO: Create vacuum job.
				fed4fireExpires: (time.Now().Add(24 * time.Hour)).Format(time.RFC3339),
				fed4fireUser:    userIdentifier.URN(),
				fed4fireSlice:   sliceIdentifier.URN(),
				fed4fireSliver:  sliverIdentifier.URN(),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					fed4fireClientId: clientId,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						fed4fireClientId: clientId,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  clientId,
							Image: image,
							Resources: corev1.ResourceRequirements{
								Limits: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU:    resource.MustParse(cpuLimit),
									corev1.ResourceMemory: resource.MustParse(memoryLimit),
								},
								Requests: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU:    resource.MustParse(defaultCpuRequest),
									corev1.ResourceMemory: resource.MustParse(defaultMemoryRequest),
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment, nil
}
