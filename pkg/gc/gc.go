package gc

import (
	"context"
	"github.com/EdgeNet-project/fed4fire/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type GC struct {
	Fed4FireClient   versioned.Interface
	KubernetesClient kubernetes.Interface
	Interval         time.Duration
	Timeout          time.Duration
	Namespace        string
}

func (w GC) Start() {
	go w.loop()
	klog.InfoS("Started collector")
}

func (w GC) loop() {
	w.collect() // Run instantly on start.
	for range time.Tick(w.Interval) {
		w.collect()
	}
}

func (w GC) collect() {
	sliversClient := w.Fed4FireClient.Fed4fireV1().Slivers(w.Namespace)
	deploymentsClient := w.KubernetesClient.AppsV1().Deployments(w.Namespace)

	ctx, cancel := context.WithTimeout(context.Background(), w.Timeout)
	defer cancel()

	slivers, err := sliversClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to list slivers")
	}

	for _, sliver := range slivers.Items {
		if time.Now().After(sliver.Spec.Expires.Time) {
			deployment, err := deploymentsClient.Get(ctx, sliver.Name, metav1.GetOptions{})
			if err != nil {
				klog.ErrorS(err, "Failed to get deployment")
				continue
			}
			err = deploymentsClient.Delete(ctx, deployment.Name, metav1.DeleteOptions{})
			if err != nil {
				klog.ErrorS(err, "Failed to delete deployment")
				continue
			}
			klog.InfoS("Deleted expired deployment", "sliver", sliver.Name)
		}
	}
}
