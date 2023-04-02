//go:build e2e

package test

import (
	"context"
	"log"
	"testing"

	"github.com/yardbirdsax/k8s-controller/controllers"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestServiceController(t *testing.T) {
	clientset, _, err := newClients(nil)
	if err != nil {
		t.Fatalf("error generating client sets: %v", err)
	}
	namespaceName := getRandomNamespace("service-controller-test")

	serviceName := "service-controller-test"
	service := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespaceName,
			Annotations: map[string]string{
				controllers.ServiceLookupAnnotationName: "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "something",
			},
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
				},
			},
		},
	}

	if _, err := clientset.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: namespaceName,
		},
	}, v1.CreateOptions{}); err != nil {
		t.Fatalf("error creating test namespace %q: %v", namespaceName, err)
	}

	if _, err := clientset.CoreV1().Services(namespaceName).Create(context.TODO(), service, v1.CreateOptions{}); err != nil {
		t.Fatalf("error creating service: %v", err)
	}

	err = wait.PollWithContext(ctx, flags.Interval, flags.Timeout, func(ctx context.Context) (done bool, err error) {
		log.Printf("[service-controller]: waiting for service to be labeled")
		service, err := clientset.CoreV1().Services(namespaceName).Get(ctx, serviceName, v1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return true, err
		}
		if val, ok := service.Labels[controllers.ServiceLabelName]; ok {
			if val == "true" {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		t.Fatal("service was not labeled within the timeout period")
	}
}
