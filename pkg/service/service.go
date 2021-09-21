package service

import "k8s.io/client-go/kubernetes"

type Service struct {
	AbsoluteURL      string
	URN              string
	KubernetesClient *kubernetes.Clientset
}
