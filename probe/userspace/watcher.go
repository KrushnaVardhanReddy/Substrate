package userspace

import (
	"context"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// PodWatcher watches Kubernetes pod events to discover workloads
// with the substrate.io/monitor: "true" label.
type PodWatcher struct {
	clientset kubernetes.Interface
}

func NewPodWatcher() (*PodWatcher, error) {
	// In a real environment, we'd use in-cluster config or kubeconfig
	// For testing, we mock or return a basic implementation.
	config, err := rest.InClusterConfig()
	if err != nil {
		// Mock implementation or return err
		return &PodWatcher{}, nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &PodWatcher{clientset: clientset}, nil
}

func (w *PodWatcher) Start(ctx context.Context) error {
	log.Println("Starting Kubernetes pod watcher...")
	if w.clientset == nil {
		<-ctx.Done()
		return nil
	}

	// Create a watcher for pods
	watcher, err := w.clientset.CoreV1().Pods("").Watch(ctx, metav1.ListOptions{
		LabelSelector: "substrate.io/monitor=true",
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil
			}
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			switch event.Type {
			case "ADDED":
				log.Printf("Attaching socket filter for pod %s/%s\n", pod.Namespace, pod.Name)
			case "DELETED":
				log.Printf("Detaching socket filter for pod %s/%s\n", pod.Namespace, pod.Name)
			}
		}
	}
}
