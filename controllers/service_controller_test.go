package controllers_test

import (
	"context"
	//"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/yardbirdsax/k8s-controller/controllers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	serviceWithLabelName    = "my-service"
	serviceWithoutLabelName = "my-other-service"
	namespaceName           = "default"
	timeout                 = 20 * time.Second
	interval                = 500 * time.Millisecond
)

var _ = Describe("Service Controller", func() {
	Context("When a service is created with the required annotation", func() {
		ctx := context.Background()

		It("should add the expected label", func() {
			service := &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						controllers.ServiceLookupAnnotationName: "true",
					},
					Name:      serviceWithLabelName,
					Namespace: namespaceName,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "something",
					},
					Type: "ClusterIP",
					Ports: []corev1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
						},
					},
				},
			}
			Eventually(func() bool {
				err := k8sClient.Create(ctx, service)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			lookupKey := types.NamespacedName{
				Namespace: namespaceName,
				Name:      serviceWithLabelName,
			}
			createdService := &corev1.Service{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdService)
				if err != nil {
					return false
				}
				labelValue, hasLabel := createdService.Labels[controllers.ServiceLabelName]
				return hasLabel && labelValue == "true"
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When a service is created without the required annotation", func() {
		ctx := context.Background()

		It("should not add the expected label", func() {
			service := &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceWithoutLabelName,
					Namespace: namespaceName,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "something",
					},
					Type: "ClusterIP",
					Ports: []corev1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
						},
					},
				},
			}
			// Setup
			Eventually(
				func() bool {
					err := k8sClient.Create(ctx, service)
					return err != nil
				},
				timeout,
				interval,
			).Should(BeTrue())
			lookupKey := types.NamespacedName{
				Namespace: namespaceName,
				Name:      serviceWithoutLabelName,
			}
			// Assert
			createdService := &corev1.Service{}
			Consistently(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdService)
				if err != nil {
					return false
				}
				_, labelExists := createdService.Labels[controllers.ServiceLabelName]
				return !labelExists
			}, timeout, interval).Should(BeTrue())
		})
	})
})
