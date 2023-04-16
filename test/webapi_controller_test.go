//go:build e2e

package test

import (
	"context"
	"log"
	"testing"

	"github.com/yardbirdsax/k8s-controller/apis/deployments/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestWebAPIController(t *testing.T) {
	t.Parallel()
	_, client, err := newClients(v1alpha1.AddToScheme)
	if err != nil {
		t.Fatalf("error generating client sets: %v", err)
	}
	namespaceName := getRandomNamespace("webapi-controller-test")

	if err = createNamespace(namespaceName, client); err != nil {
		t.Fatalf("error creating namespace %q: %v", namespaceName, err)
	}

	webAPIName := "web-api"
	webAPI := &v1alpha1.WebAPI{
		ObjectMeta: v1.ObjectMeta{
			Name:      webAPIName,
			Namespace: namespaceName,
		},
		Spec: v1alpha1.WebAPISpec{},
	}

	if err = client.Create(ctx, webAPI); err != nil {
		t.Fatalf("error creating web API resource: %v", err)
	}

	err = wait.PollWithContext(ctx, flags.Interval, flags.Timeout, func(ctx context.Context) (done bool, err error) {
		log.Printf("[web-api-controller]: waiting for web-api to be labeled")
		err = client.Get(ctx, types.NamespacedName{Namespace: namespaceName, Name: webAPIName}, webAPI)
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return true, err
		}
		if val, ok := webAPI.Labels["labeled"]; ok {
			if val == "true" {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		t.Fatal("web API was not labeled within the timeout period")
	}
}
