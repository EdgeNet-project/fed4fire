package service

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"net/http"
)

type AllocateArgs struct {
	SliceURN    string
	Credentials []Credential
	Rspec       string
	Options     Options
}

type AllocateReply struct {
	Data struct {
		Code  Code   `xml:"code"`
		Value string `xml:"value"`
	}
}

func (s *Service) Allocate(r *http.Request, args *AllocateArgs, reply *AllocateReply) error {
	fmt.Println(args.Rspec)
	fmt.Println(args.Options)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "todo",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "todo",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "todo",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "todo",
							Image: "docker.io/library/ubuntu:20.04",
							// TODO: Port?
						},
					},
				},
			},
		},
	}

	deploymentsClient:= s.KubernetesClient.AppsV1().Deployments("fed4fire-todo")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(result)

	//reply.Data.Code.Code = geniCodeSuccess
	//xml_, err := xml.Marshal(v)
	//if err != nil {
	//	fmt.Println(err)
	//	return err
	//}
	//reply.Data.Value = string(xml_)
	return nil
}
