//go:build e2e

package test

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type testFlags struct {
	Kubeconfig string
	Cluster    string
	Timeout    time.Duration
	Interval   time.Duration
}

var flags = initFlags()
var ctx context.Context

/*
Many of the patterns here were inspired by (or in some cases directly sourced from) the work
by Christie Wilson on testing CRDs: https://github.com/bobcatfish/testing-crds. Check out
that repository for more great ideas!
*/

func initFlags() *testFlags {
	log.Println("starting init flags")
	flag.VisitAll(func(f *flag.Flag) {
		log.Printf("value for flag %q is %q", f.Name, f.Value)
	})
	f := testFlags{}
	currentUser, _ := user.Current()
	defaultKubeconfig := filepath.Join(currentUser.HomeDir, ".kube/config")
	f.Kubeconfig = flag.Lookup("kubeconfig").Value.String()
	if f.Kubeconfig == "" {
		f.Kubeconfig = defaultKubeconfig
	}
	flag.StringVar(&f.Cluster, "cluster", "", "The name of the cluster for the Kubeconfig context to use.")
	flag.DurationVar(&f.Timeout, "timeout", 5*time.Minute, "The maxiumum duration of test execution.")
	flag.DurationVar(&f.Interval, "interval", 1*time.Second, "The interval between polls for checking the status of resources.")
	return &f
}

type schemeAdder func(*runtime.Scheme) error

func newClients(schemeAdder schemeAdder) (*kubernetes.Clientset, client.Client, error) {
	if schemeAdder != nil {
		schemeAdder(scheme.Scheme)
	}

	overrides := clientcmd.ConfigOverrides{}
	if flags.Cluster != "" {
		overrides.Context.Cluster = flags.Cluster
	}
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: flags.Kubeconfig,
		},
		&overrides,
	)
	clientConfig, err := cfg.ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting client config with kubeconfig %q and cluster %q: %w", flags.Kubeconfig, flags.Cluster, err)
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating clientset from client config (%+v): %w", *clientConfig, err)
	}
	client, err := client.New(clientConfig, client.Options{})

	return clientset, client, nil
}

func getRandomNamespace(prefix string) string {
	rand.Seed(time.Now().Unix())
	suffix := make([]string, len(prefix))
	for range prefix {
		// 97 - 122 represent the ascii characters 'a-z'
		suffix = append(suffix, fmt.Sprintf("%c", rand.Intn(122-97)+97))
	}
	return strings.Join([]string{prefix, strings.Join(suffix, "")}, "-")
}

func createNamespace(name string, client client.Client) error {
	namespace := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}
	return client.Create(ctx, namespace)
}
