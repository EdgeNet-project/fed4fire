package gc

import (
	"context"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type GC struct {
	Client    kubernetes.Interface
	Interval  time.Duration
	Namespace string
}

func (gc GC) Start() {
	go gc.loop()
}

func (gc GC) loop() {
	for range time.Tick(gc.Interval) {
		gc.collect()
	}
}

func (gc GC) collect() {
	client := gc.Client.AppsV1().Deployments(gc.Namespace)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	deployments, err := client.List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to list deployments")
	}
	for _, deployment := range deployments.Items {
		expirationTimeStr := deployment.Annotations[constants.Fed4FireExpires]
		expirationTime, err := time.Parse(time.RFC3339, expirationTimeStr)
		if err != nil {
			klog.ErrorS(err, "Failed to parse expiration time", "value", expirationTimeStr)
		}
		if expirationTime.Before(time.Now()) {
			err := client.Delete(ctx, deployment.Name, metav1.DeleteOptions{})
			if err == nil {
				klog.InfoS("Deleted deployment", "name", deployment.Name)
			} else {
				klog.ErrorS(err, "Failed to delete deployment", "name", deployment.Name)
			}
		}
	}
}
