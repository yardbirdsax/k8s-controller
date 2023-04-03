package deployments

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	"github.com/yardbirdsax/k8s-controller/apis/deployments/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestLabel(t *testing.T) {
	webapi := &v1alpha1.WebAPI{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{},
		},
	}

	label(webapi)

	require.Contains(t, webapi.ObjectMeta.Labels, "labeled", "labels does not contain expected key")
	require.Equal(t, "true", webapi.Labels["labeled"], "label is not expected value")
}

var _ = Describe("WebAPI Controller", func() {
	const (
		timeout    = time.Second * 30
		duration   = time.Second * 30
		interval   = time.Second * 5
		webAPIName = "web-api"
		namespace  = "default"
	)

	Context("When a new WebAPI resource is created", func() {
		It("should label the resource", func() {
			ctx := context.Background()
			webAPI := &v1alpha1.WebAPI{
				ObjectMeta: v1.ObjectMeta{
					Name:      webAPIName,
					Namespace: namespace,
				},
				Spec: v1alpha1.WebAPISpec{
					Foo: "bar",
				},
			}
			Expect(k8sClient.Create(ctx, webAPI)).Should(Succeed())

			webAPILookupKey := types.NamespacedName{Name: webAPIName, Namespace: namespace}
			createdWebAPI := &v1alpha1.WebAPI{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, webAPILookupKey, createdWebAPI)
				if err != nil {
					return false
				}
				if createdWebAPI.Labels["labeled"] == "true" {
					return true
				}
				return false
			}, timeout).Should(BeTrue())
		})
	})
})
