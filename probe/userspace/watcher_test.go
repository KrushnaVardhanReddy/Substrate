package userspace

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPodWatcher(t *testing.T) {
	// Create fake clientset
	clientset := fake.NewSimpleClientset()

	watcher := &PodWatcher{
		clientset: clientset,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start watcher in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- watcher.Start(ctx)
	}()

	// Wait a bit for the watcher to start
	time.Sleep(100 * time.Millisecond)

	// Create a pod with the label
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels: map[string]string{
				"substrate.io/monitor": "true",
			},
		},
	}

	_, err := clientset.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Delete the pod
	err = clientset.CoreV1().Pods("default").Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
	assert.NoError(t, err)

	// Wait for the context to timeout
	err = <-errCh
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
