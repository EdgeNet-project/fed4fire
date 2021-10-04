package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/edgenet/pkg/apis/core/v1alpha"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	"github.com/EdgeNet-project/fed4fire/pkg/urn"
	"html"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"net/http"
	"strings"
)

type AllocateArgs struct {
	SliceURN    string
	Credentials []Credential
	Rspec       string
	Options     string // Options
}

type AllocateReply struct {
	Data struct {
		Code  Code   `xml:"code"`
		Value string `xml:"value"`
	}
}

// Allocate resources as described in a request RSpec argument to a slice with the named URN.
// On success, one or more slivers are allocated, containing resources satisfying the request, and assigned to the given slice.
// This method returns a listing and description of the resources reserved for the slice by this operation, in the form of a manifest RSpec.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Allocate
func (s *Service) Allocate(r *http.Request, args *AllocateArgs, reply *AllocateReply) error {
	credential := args.Credentials[0].SFA().Credential

	v := rspec.Rspec{}
	a := []byte(html.UnescapeString(args.Rspec))
	err := xml.Unmarshal(a, &v)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(v)
	//fmt.Println(v.Nodes[0])

	// 1. Create a subnamespace for the slice
	sliceIdentifier, err := urn.Parse(args.SliceURN)
	if err != nil {
		reply.Data.Value = "Failed to parse slice URN"
		reply.Data.Code.Code = geniCodeError
		klog.ErrorS(err, reply.Data.Value, "urn", args.SliceURN)
		return nil
	}
	subnamespaceName, err := subnamespaceNameFor(*sliceIdentifier)
	if err != nil {
		reply.Data.Value = "Failed to build subnamespace name"
		reply.Data.Code.Code = geniCodeError
		klog.ErrorS(err, reply.Data.Value, "identifier", sliceIdentifier)
		return nil
	}
	subnamespace, err := s.EdgenetClient.CoreV1alpha().SubNamespaces(s.ParentNamespace).Get(context.TODO(), *subnamespaceName, metav1.GetOptions{})
	if err != nil {
		klog.InfoS(
			"Could not find subnamespace", "subnamespace", *subnamespaceName,
		)
		subnamespace = &v1alpha.SubNamespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: *subnamespaceName,
				Annotations: map[string]string{
					fed4fireAnnotationSlice: args.SliceURN,
				},
			},
			Spec: v1alpha.SubNamespaceSpec{
				Resources: v1alpha.Resources{
					CPU:    s.NamespaceCpuLimit,
					Memory: s.NamespaceMemoryLimit,
				},
				Inheritance: v1alpha.Inheritance{
					NetworkPolicy: true,
					RBAC:          true,
				},
			},
		}
		_, err = s.EdgenetClient.CoreV1alpha().SubNamespaces(s.ParentNamespace).Create(context.TODO(), subnamespace, metav1.CreateOptions{})
		if err != nil {
			reply.Data.Value = "Failed to create subnamespace"
			reply.Data.Code.Code = geniCodeError
			klog.ErrorS(err, reply.Data.Value, "subnamespace", *subnamespace)
		}
		klog.InfoS("Created subnamespace", "subnamespace", *subnamespaceName)
	}

	deployments := make([]appsv1.Deployment, 0)
	for _, node := range v.Nodes {
		deployments = append(deployments, appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: node.ClientID,
				Annotations: map[string]string{
					// TODO: Are multiple sliver types allowed?
					// If not should we validate agains the schema before?
					// TODO: Validate requested image name, and use default if not specified.
					// TODO: Create vacuum job.
					fed4fireExpiryTime: (time.Now().Add(86400 * time.Second)).String(),
					fed4fireImageName:  "todo",
					fed4fireSlice:      args.SliceURN,
					fed4fireUser:       credential.TargetURN,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: pointer.Int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						fed4fireClientId: node.ClientID,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							fed4fireClientId: node.ClientID,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  node.ClientID,
								Image: defaultPauseImage,
								Resources: corev1.ResourceRequirements{
									Limits: map[corev1.ResourceName]resource.Quantity{
										corev1.ResourceCPU:    resource.MustParse(s.ContainerCpuLimit),
										corev1.ResourceMemory: resource.MustParse(s.ContainerMemoryLimit),
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
		})
	}

	deploymentsClient := s.KubernetesClient.AppsV1().Deployments(fmt.Sprintf("%s-%s", s.ParentNamespace, *subnamespaceName))
	deployed := make([]appsv1.Deployment, 0)
	success := true
	for _, deployment := range deployments {
		_, err := deploymentsClient.Create(context.TODO(), &deployment, metav1.CreateOptions{})
		if err != nil {
			klog.ErrorS(err, "Failed to create deployment", "deployment", deployment.Name)
			success = false
			break
		}
		deployed = append(deployed, deployment)
	}
	if !success {
		for _, deployment := range deployed {
			klog.InfoS("Rolling back deployment", "deployment", deployment)
			err = deploymentsClient.Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{})
			if err != nil {
				klog.InfoS("Failed to delete deployment", "deployment", deployment.Name)
			}
		}
		reply.Data.Value = "Failed to create deployment(s)"
		reply.Data.Code.Code = geniCodeError
		klog.ErrorS(err, reply.Data.Value)
		return nil
	}

	// TODO: Ignore already existing resources, and return slivers (only newly allocated ones).
	// TODO: Implement expiration using labels?

	reply.Data.Code.Code = geniCodeSuccess
	return nil
}

func subnamespaceNameFor(identifier urn.Identifier) (*string, error) {
	if identifier.ResourceType != "slice" {
		return nil, fmt.Errorf("URN resource type must be `slice`")
	}
	s := fmt.Sprintf(
		"%s-%s",
		strings.Join(identifier.Authorities, "-"),
		identifier.ResourceName,
	)
	return &s, nil
}
