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
	fmt.Println(string(a))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(v)
	fmt.Println(v.Nodes[0])

	// 1. Create a subnamespace for the slice
	sliceIdentifier, err := urn.Parse(args.SliceURN)
	if err != nil {
		klog.ErrorS(err, "Failed to parse slice URN", "urn", args.SliceURN)
		reply.Data.Code.Code = geniCodeError
		reply.Data.Value = "Failed to parse slice URN"
		return nil
	}
	subnamespaceName, err := subnamespaceNameFor(*sliceIdentifier)
	if err != nil {
		klog.ErrorS(err, "Failed to build subnamespace name", "identifier", sliceIdentifier)
		reply.Data.Code.Code = geniCodeError
		reply.Data.Value = "Failed to build subnamespace name"
		return nil
	}
	subnamespace, err := s.EdgenetClient.CoreV1alpha().SubNamespaces("lip6-lab-fed4fire-dev").Get(context.TODO(), *subnamespaceName, metav1.GetOptions{})
	if err != nil {
		klog.InfoS(
			"Could not find subnamespace", "subnamespace", *subnamespaceName,
		)
		subnamespace = &v1alpha.SubNamespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: *subnamespaceName,
				Annotations: map[string]string{
					annotationSlice: args.SliceURN,
				},
			},
			Spec: v1alpha.SubNamespaceSpec{
				Resources: v1alpha.Resources{
					CPU:    "8",
					Memory: "8Gi",
				},
				Inheritance: v1alpha.Inheritance{
					NetworkPolicy: true,
					RBAC:          true,
				},
			},
		}
		_, err = s.EdgenetClient.CoreV1alpha().SubNamespaces("lip6-lab-fed4fire-dev").Create(context.TODO(), subnamespace, metav1.CreateOptions{})
		if err != nil {
			klog.ErrorS(err, "Failed to create subnamespace", subnamespace, *subnamespace)
			reply.Data.Code.Code = geniCodeError
			reply.Data.Value = "Failed to create subnamespace"
		}
		klog.InfoS("Created subnamespace", subnamespace, *subnamespaceName)
	}

	deployments := make([]appsv1.Deployment, 0)
	for _, node := range v.Nodes {
		deployments = append(deployments, appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: node.ClientID,
				Annotations: map[string]string{
					annotationSlice: args.SliceURN,
					annotationUser:  credential.TargetURN,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: pointer.Int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": node.ClientID,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": node.ClientID,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  node.ClientID,
								Image: "docker.io/library/ubuntu:20.04",
								// TODO: Port?
								Resources: corev1.ResourceRequirements{
									Limits: map[corev1.ResourceName]resource.Quantity{
										corev1.ResourceCPU:    resource.MustParse("2"),
										corev1.ResourceMemory: resource.MustParse("2Gi"),
									},
									Requests: map[corev1.ResourceName]resource.Quantity{
										corev1.ResourceCPU:    resource.MustParse("0.1"),
										corev1.ResourceMemory: resource.MustParse("128Mi"),
									},
								},
							},
						},
					},
				},
			},
		})
	}

	deploymentsClient := s.KubernetesClient.AppsV1().Deployments("lip6-lab-fed4fire-dev")
	for _, deployment := range deployments {
		_, err := deploymentsClient.Create(context.TODO(), &deployment, metav1.CreateOptions{})
		fmt.Println(err)
	}
	//result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	//fmt.Println(err)
	//if err != nil {
	//	fmt.Println(err)
	//	return err
	//}
	//fmt.Println(result)

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
