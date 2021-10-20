package service

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/naming"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"

	"k8s.io/apimachinery/pkg/api/errors"

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

func (v *AllocateReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) error {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = geniCodeError
	v.Data.Value.Error = fmt.Sprintf("%s: %s", msg, err)
	return nil
}

// Some things to take into account in request RSpecs:
// - Each node will have exactly one sliver_type in a request.
// - Each sliver_type will have zero or one disk_image elements.
//   If your testbed requires disk_image or does not support it,
//   it should handle bad requests RSpecs with the correct error.
// - The exclusive element is specified for each node in the request.
//   Your testbed should check if the specified value (in combination with the sliver_type) is supported,
//   and return the correct error if not.
// - The request RSpec might contain links that have a component_manager element that matches your AM.
//   If your AM does not support links, it should return the correct error.
// https://doc.fed4fire.eu/testbed_owner/rspec.html#request-rspec

// Some information will be in a request RSpec, that needs to be ignored and copied to the manifest RSpec unaltered.
// This is important to do correctly.
// - A request RSpec can contain nodes that have a component_manager_id set to a different AM.
//   You need to ignore these nodes, and copy them to the manifest RSpec unaltered.
// - A request RSpec can contain links that do not have a component_manager matching your AM
//   (links have multiple component_manager_id elements!).
//   You need to ignore these links, and copy them to the manifest RSpec unaltered.
// - A request RSpec can contain XML extensions in nodes, links, services, or directly in the rspec element.
//   Some of these the AM will not know.
//   It has to ignore these, and preferably also pass them unaltered to the manifest RSpec.
// https://doc.fed4fire.eu/testbed_owner/rspec.html#request-rspec

// Allocate allocates resources as described in a request RSpec argument to a slice with the named URN.
// On success, one or more slivers are allocated, containing resources satisfying the request, and assigned to the given slice.
// This method returns a listing and description of the resources reserved for the slice by this operation, in the form of a manifest RSpec.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Allocate
func (s *Service) Allocate(r *http.Request, args *AllocateArgs, reply *AllocateReply) error {
	// Allocate moves 1 or more slivers from geni_unallocated (state 1) to geni_allocated (state 2).
	// This method can be described as creating an instance of the state machine for each sliver.
	// If the aggregate cannot fully satisfy the request, the whole request fails.
	// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3/CommonConcepts#SliverAllocationStates
	userIdentifier, err := identifiers.Parse(r.Header.Get(utils.HttpHeaderUser))
	if err != nil {
		return reply.SetAndLogError(err, "Failed to parse user URN")
	}
	sliceIdentifier, err := identifiers.Parse(args.SliceURN)
	if err != nil {
		return reply.SetAndLogError(err, "Failed to parse slice URN")
	}
	credential, err := FindMatchingCredential(
		*userIdentifier,
		*sliceIdentifier,
		args.Credentials,
		s.TrustedCertificates,
	)
	if err == nil {
		klog.InfoS(
			"Found matching credential",
			"ownerUrn",
			credential.OwnerURN,
			"targetUrn",
			credential.TargetURN,
		)
	} else {
		return reply.SetAndLogError(err, "Invalid credentials")
	}

	// TODO: Implement RSpec passthroughs + node selection.
	requestRspec := rspec.Rspec{}
	err = xml.Unmarshal([]byte(html.UnescapeString(args.Rspec)), &requestRspec)
	if err != nil {
		return reply.SetAndLogError(err, "Failed to deserialize rspec")
	}

	// Build the deployment and service objects
	deployments := make([]*appsv1.Deployment, len(requestRspec.Nodes))
	services := make([]*corev1.Service, len(requestRspec.Nodes))
	for i, node := range requestRspec.Nodes {
		deployments[i], err = deploymentForRspec(
			node,
			s.AuthorityIdentifier,
			*sliceIdentifier,
			*userIdentifier,
			s.ContainerImages,
			s.ContainerCpuLimit,
			s.ContainerMemoryLimit,
		)
		if err != nil {
			return reply.SetAndLogError(err, "Failed to build deployment")
		}
		services[i], err = serviceForRspec(node, *sliceIdentifier)
		if err != nil {
			return reply.SetAndLogError(err, "Failed to build service")
		}
	}

	// Create the deployment and service objects and rollback them in case of failure
	// TODO: Duplicate deployed for deployments and services for proper rollback.
	// Or use tuple?
	deployed := make([]bool, len(deployments))
	success := true
	for i, deployment := range deployments {
		_, err := s.Deployments().Create(r.Context(), deployment, metav1.CreateOptions{})
		if err == nil {
			klog.InfoS("Created deployment", "name", deployment.Name)
			deployed[i] = true
		} else if errors.IsAlreadyExists(err) {
			klog.InfoS("Ignoring already existing deployment", "name", deployment.Name)
			deployed[i] = false
		} else {
			_ = reply.SetAndLogError(err, "Failed to create deployment", "name", deployment.Name)
			deployed[i] = false
			success = false
			break
		}
	}
	// Create the services
	for i, service := range services {
		_, err := s.Services().Create(r.Context(), service, metav1.CreateOptions{})
		if err == nil {
			klog.InfoS("Created service", "name", service.Name)
			deployed[i] = true
		} else if errors.IsAlreadyExists(err) {
			klog.InfoS("Ignoring already existing service", "name", service.Name)
			deployed[i] = false
		} else {
			_ = reply.SetAndLogError(err, "Failed to create service", "name", service.Name)
			deployed[i] = false
			success = false
			break
		}
	}
	// Rollback in case of failure
	if !success {
		for i, isDeployed := range deployed {
			if isDeployed {
				deployment := deployments[i]
				service := services[i]
				err = s.Deployments().Delete(r.Context(), deployment.Name, metav1.DeleteOptions{})
				if err == nil {
					klog.InfoS("Deleted deployment", "name", deployment.Name)
				} else {
					klog.InfoS("Failed to delete deployment", "name", deployment.Name)
				}
				err = s.Services().Delete(r.Context(), service.Name, metav1.DeleteOptions{})
				if err == nil {
					klog.InfoS("Deleted service", "name", service.Name)
				} else {
					klog.InfoS("Failed to delete service", "name", service.Name)
				}
			}
		}
		return nil
	}

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
		return reply.SetAndLogError(err, "Failed to serialize response")
	}
	reply.Data.Value.Rspec = string(xml_)
	reply.Data.Code.Code = geniCodeSuccess
	return nil
}

func deploymentImageForSliverType(
	sliverType rspec.SliverType,
	containerImages map[string]string,
) (string, error) {
	if len(sliverType.DiskImages) == 0 {
		return containerImages[utils.Keys(containerImages)[0]], nil
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
	sliceHash := naming.SliceHash(sliceIdentifier)
	sliverName, err := naming.SliverName(sliceIdentifier, node.ClientID)
	if err != nil {
		return nil, err
	}
	sliverIdentifier := authorityIdentifier.Copy(identifiers.ResourceTypeSliver, sliverName)
	annotations := map[string]string{
		fed4fireClientId: node.ClientID,
		// TODO: Create vacuum job.
		fed4fireExpires: (time.Now().Add(24 * time.Hour)).Format(time.RFC3339),
		fed4fireUser:    userIdentifier.URN(),
		fed4fireSlice:   sliceIdentifier.URN(),
		fed4fireSliver:  sliverIdentifier.URN(),
	}
	labels := map[string]string{
		fed4fireSliceHash:  sliceHash,
		fed4fireSliverName: sliverName,
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        sliverName,
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					fed4fireSliverName: sliverName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  sliverName,
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

func serviceForRspec(
	node rspec.Node,
	sliceIdentifier identifiers.Identifier,
) (*corev1.Service, error) {
	sliceHash := naming.SliceHash(sliceIdentifier)
	sliverName, err := naming.SliverName(sliceIdentifier, node.ClientID)
	if err != nil {
		return nil, err
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: sliverName,
			Labels: map[string]string{
				fed4fireSliceHash:  sliceHash,
				fed4fireSliverName: sliverName,
			},
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{Port: 22}},
			Selector: map[string]string{
				fed4fireSliverName: sliverName,
			},
		},
	}
	return service, nil
}
