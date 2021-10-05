package service

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/EdgeNet-project/edgenet/pkg/apis/core/v1alpha"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	"github.com/EdgeNet-project/fed4fire/pkg/urn"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

type Sliver struct {
	URN              string `xml:"geni_sliver_urn"`
	Expires          string `xml:"geni_expires"`
	AllocationStatus string `xml:"geni_allocation_status"`
}

type AllocateArgs struct {
	SliceURN    string
	Credentials []Credential
	Rspec       string
	Options     string // Options
}

type AllocateReply struct {
	Data struct {
		Code  Code `xml:"code"`
		Value struct {
			Rspec   string   `xml:"geni_rspec"`
			Slivers []Sliver `xml:"geni_slivers"`
		} `xml:"value"`
	}
}

func (v *AllocateReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = geniCodeError
	// TODO
	// v.Data.Value = fmt.Sprintf("%s: %s", msg, err)
}

// Allocate allocates resources as described in a request RSpec argument to a slice with the named URN.
// On success, one or more slivers are allocated, containing resources satisfying the request, and assigned to the given slice.
// This method returns a listing and description of the resources reserved for the slice by this operation, in the form of a manifest RSpec.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Allocate
func (s *Service) Allocate(r *http.Request, args *AllocateArgs, reply *AllocateReply) error {
	credential := args.Credentials[0].SFA().Credential

	v := rspec.Rspec{}
	err := xml.Unmarshal([]byte(html.UnescapeString(args.Rspec)), &v)
	if err != nil {
		reply.SetAndLogError(err, "Failed to deserialize rspec")
		return nil
	}

	// 1. Get or create the subnamespace for the slice
	subnamespaceClient := s.EdgenetClient.CoreV1alpha().SubNamespaces(s.ParentNamespace)
	subnamespaceName, err := subnamespaceNameForSlice(args.SliceURN)
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
			args.SliceURN,
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
	deployments := make([]appsv1.Deployment, 0)
	for _, node := range v.Nodes {
		deployment, err := deploymentForRspec(
			node,
			args.SliceURN,
			credential.TargetURN,
			s.ContainerImages,
			s.ContainerCpuLimit,
			s.ContainerMemoryLimit,
		)
		if err != nil {
			reply.SetAndLogError(err, "Failed to build deployment")
			return nil
		}
		deployments = append(deployments, *deployment)
	}

	// 3. Create the deployment objects and rollback them in case of failure
	targetNamespace := fmt.Sprintf("%s-%s", s.ParentNamespace, subnamespaceName)
	deploymentsClient := s.KubernetesClient.AppsV1().Deployments(targetNamespace)
	deployed := make([]appsv1.Deployment, 0)
	success := true
	for _, deployment := range deployments {
		_, err := deploymentsClient.Create(r.Context(), &deployment, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			klog.InfoS("Ignoring already existing deployment", "name", deployment.Name)
		} else if err != nil {
			reply.SetAndLogError(err, "Failed to create deployment", "name", deployment.Name)
			success = false
			break
		} else {
			deployed = append(deployed, deployment)
		}
	}
	if !success {
		for _, deployment := range deployed {
			klog.InfoS("Rolling back deployment", "name", deployment)
			err = deploymentsClient.Delete(r.Context(), deployment.Name, metav1.DeleteOptions{})
			if err != nil {
				klog.InfoS("Failed to rollback deployment", "name", deployment.Name)
			}
		}
		return nil
	}

	// TODO: Ignore already existing resources, and return slivers (only newly allocated ones).
	reply.Data.Value.Rspec = html.UnescapeString(args.Rspec)
	// TODO: Proper values
	reply.Data.Value.Slivers = append(reply.Data.Value.Slivers, Sliver{
		URN:              "test",
		Expires:          "test",
		AllocationStatus: geniStateAllocated,
	})

	reply.Data.Code.Code = geniCodeSuccess
	return nil
}

func subnamespaceNameForSlice(sliceUrn string) (string, error) {
	sliceIdentifier, err := urn.Parse(sliceUrn)
	if err != nil {
		return "", err
	}
	if sliceIdentifier.ResourceType != "slice" {
		return "", fmt.Errorf("URN resource type must be `slice`")
	}
	s := fmt.Sprintf(
		"fed4fire-slice-%s-%s",
		strings.Join(sliceIdentifier.Authorities, "-"),
		sliceIdentifier.ResourceName,
	)
	s = strings.ReplaceAll(s, ".", "-")
	return s, nil
}

func subnamespaceForSlice(
	sliceUrn string,
	cpuLimit string,
	memoryLimit string,
) (*v1alpha.SubNamespace, error) {
	name, err := subnamespaceNameForSlice(sliceUrn)
	if err != nil {
		return nil, err
	}
	subnamespace := &v1alpha.SubNamespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				fed4fireSlice: sliceUrn,
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
		for _, image := range containerImages {
			return image, nil
		}
	}
	if len(sliverType.DiskImages) == 1 {
		identifier, err := urn.Parse(sliverType.DiskImages[0].Name)
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
	sliceUrn string,
	userUrn string,
	containerImages map[string]string,
	cpuLimit string,
	memoryLimit string,
) (*appsv1.Deployment, error) {
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
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: clientId,
			Annotations: map[string]string{
				// TODO: Create vacuum job.
				fed4fireExpiryTime: (time.Now().Add(86400 * time.Second)).String(),
				fed4fireImage:      image,
				fed4fireSlice:      sliceUrn,
				fed4fireUser:       userUrn,
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
							Image: defaultPauseImage,
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
