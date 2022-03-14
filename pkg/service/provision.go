package service

import (
	"context"
	"encoding/xml"
	"fmt"
	v1 "github.com/EdgeNet-project/fed4fire/pkg/apis/fed4fire/v1"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/EdgeNet-project/fed4fire/pkg/naming"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"net/http"
	"strings"
)

type ProvisionArgs struct {
	URNs        []string
	Credentials []Credential
	Options     Options
}

type ProvisionReply struct {
	Data struct {
		Code   Code   `xml:"code"`
		Output string `xml:"output"`
		Value  struct {
			Rspec   string   `xml:"geni_rspec"`
			Slivers []Sliver `xml:"geni_slivers"`
		} `xml:"value"`
	}
}

func (v *ProvisionReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) error {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Output = fmt.Sprintf("%s: %s", msg, err)
	return nil
}

// Provision requests that the named geni_allocated slivers be made geni_provisioned,
// instantiating or otherwise realizing the resources, such that they have a valid geni_operational_status
// and may be made geni_ready for experimenter use.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Provision
func (s *Service) Provision(r *http.Request, args *ProvisionArgs, reply *ProvisionReply) error {
	slivers, err := s.AuthorizeAndListSlivers(
		r.Context(),
		r.Header.Get(constants.HttpHeaderUser),
		args.URNs,
		args.Credentials,
	)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorListResources)
	}

	sshKeys := make([]string, 0)
	for _, user := range args.Options.Users {
		for _, key := range user.Keys {
			sshKeys = append(sshKeys, key)
		}
	}

	// Build the sliver resources
	resources := make([]*sliverResources, len(slivers))
	for i, sliver := range slivers {
		resources[i], err = buildResources(
			sliver,
			sshKeys,
			s.ContainerCpuLimit,
			s.ContainerMemoryLimit,
		)
		if err != nil {
			return reply.SetAndLogError(err, constants.ErrorBuildResources)
		}
	}

	// Create the sliver resources and roll them back in case of failure
	var createResourcesError error
	for _, res := range resources {
		createResourcesError = createResources(r.Context(), *s, *res)
		if createResourcesError != nil {
			break
		}
	}

	// Rollback in case of failure
	if createResourcesError != nil {
		for _, res := range resources {
			if deleteResources(r.Context(), *s, *res) != nil {
				klog.InfoS("Failed to delete resources", "name", res.Deployment.Name)
			}
		}
		return reply.SetAndLogError(createResourcesError, constants.ErrorCreateResource)
	}

	returnRspec := rspec.Rspec{Type: rspec.RspecTypeManifest}

	for _, sliver := range slivers {

		sliver, err := s.Slivers().Get(r.Context(), sliver.Name, metav1.GetOptions{})
		if err != nil {
			return reply.SetAndLogError(err, constants.ErrorGetResource)
		}
		allocationStatus, operationalStatus := s.GetSliverStatus(r.Context(), sliver.Name)
		reply.Data.Value.Slivers = append(
			reply.Data.Value.Slivers,
			NewSliver(*sliver, allocationStatus, operationalStatus),
		)
		returnRspec.Nodes = append(returnRspec.Nodes, rspec.Node{
			ComponentManagerID: s.AuthorityIdentifier.URN(),
			Available:          rspec.Available{Now: false},
			ClientID:           sliver.Spec.ClientID,
			Exclusive:          false,
		})
	}

	xml_, err := xml.Marshal(returnRspec)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorSerializeRspec)
	}
	reply.Data.Value.Rspec = string(xml_)
	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}

type sliverResources struct {
	ConfigMap  *corev1.ConfigMap
	Deployment *appsv1.Deployment
	Service    *corev1.Service
}

func buildResources(
	sliver v1.Sliver,
	sshKeys []string,
	cpuLimit string,
	memoryLimit string,
) (*sliverResources, error) {
	labels := map[string]string{
		constants.Fed4FireSliceHash:  naming.SliceHash(sliver.Spec.SliceURN),
		constants.Fed4FireSliverName: sliver.Name,
	}

	nodeSelectorRequirements := []corev1.NodeSelectorRequirement{{
		Key:      corev1.LabelOSStable,
		Operator: corev1.NodeSelectorOpIn,
		Values:   []string{"linux"},
	}}
	if sliver.Spec.RequestedArch != nil {
		nodeSelectorRequirements = append(nodeSelectorRequirements, corev1.NodeSelectorRequirement{
			Key:      corev1.LabelArchStable,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{*sliver.Spec.RequestedArch},
		})
	}
	if sliver.Spec.RequestedNode != nil {
		nodeSelectorRequirements = append(nodeSelectorRequirements, corev1.NodeSelectorRequirement{
			Key:      corev1.LabelHostname,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{*sliver.Spec.RequestedNode},
		})
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   sliver.Name,
			Labels: labels,
		},
		Data: map[string]string{
			"authorized_keys": strings.Join(sshKeys, "\n") + "\n",
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   sliver.Name,
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constants.Fed4FireSliverName: sliver.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{{
									MatchExpressions: nodeSelectorRequirements,
								}},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  sliver.Name,
							Image: sliver.Spec.Image,
							Resources: corev1.ResourceRequirements{
								Limits: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU:    resource.MustParse(cpuLimit),
									corev1.ResourceMemory: resource.MustParse(memoryLimit),
								},
								Requests: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU: resource.MustParse(
										constants.DefaultCpuRequest,
									),
									corev1.ResourceMemory: resource.MustParse(
										constants.DefaultMemoryRequest,
									),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "ssh-volume",
									ReadOnly:  true,
									MountPath: "/root/.ssh/authorized_keys",
									SubPath:   "authorized_keys",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "ssh-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMap.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   sliver.Name,
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{Port: 22},
			},
			Selector: map[string]string{
				constants.Fed4FireSliverName: sliver.Name,
			},
		},
	}

	return &sliverResources{configMap, deployment, service}, nil
}

func createResources(context context.Context, service Service, resources sliverResources) error {
	sliver, err := service.Slivers().Get(context, resources.Deployment.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	ownerReference := metav1.OwnerReference{
		APIVersion: "fed4fire.edgenet.io/v1",
		Kind:       "Sliver",
		Name:       sliver.Name,
		UID:        sliver.UID,
	}
	resources.Deployment.OwnerReferences = append(
		resources.Deployment.OwnerReferences,
		ownerReference,
	)
	_, err = service.Deployments().Create(context, resources.Deployment, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	resources.ConfigMap.OwnerReferences = append(
		resources.ConfigMap.OwnerReferences,
		ownerReference,
	)
	_, err = service.ConfigMaps().Create(context, resources.ConfigMap, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	resources.Service.OwnerReferences = append(resources.Service.OwnerReferences, ownerReference)
	_, err = service.Services().Create(context, resources.Service, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func deleteResources(context context.Context, service Service, resources sliverResources) error {
	err := service.Services().Delete(context, resources.Service.Name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	err = service.ConfigMaps().Delete(context, resources.ConfigMap.Name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	err = service.Deployments().Delete(context, resources.Deployment.Name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}
