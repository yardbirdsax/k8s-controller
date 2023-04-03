
# Image URL to use all building/pushing image targets
IMG ?= k8s-controller
IMG_TAG ?= latest
IMG_REGISTRY ?=
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.24.1
# The name of the local k3d cluster for testing
K3D_CLUSTER_NAME=$(shell cat cluster.yaml| yq 'select(.k3d != null) | .k3d.v1alpha4Simple.metadata.name')
# The name of the Helm release when installing using Helm
HELM_RELEASE_NAME ?= "k8s-controller"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
.PHONY: helm
helm: manifests kustomize helmify ## Generate Helm charts
	$(KUSTOMIZE) build config/default | $(HELMIFY) charts/k8s-controller

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	make start
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" \
	KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT=true \
	USE_EXISTING_CLUSTER=true go test ./... -coverprofile cover.out -test.gocoverdir "$$PWD/coverage/unit"; \
	make stop

.PHONY: test-e2e
test-e2e: start docker-build docker-load deploy
	go test -v -tags=e2e -count=1 ./test --timeout 30s; \
	if [ $$? -eq 0 ]; then \
		echo "test succeeded"; \
	else \
		echo "test failed!"; \
	fi;
	make stop

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG}:${IMG_TAG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG_REGISTRY}/${IMG}:${IMG_TAG}

.PHONY: docker-tag
docker-tag: ## Tag docker image with a new registry URL attached
	docker tag ${IMG}:${IMG_TAG} ${IMG_REGISTRY}/${IMG}:${IMG_TAG}

.PHONY: docker-load
docker-load: ## Loads the image onto the local k3d cluster
	$(K3D) image load ${IMG}:${IMG_TAG} -c $(K3D_CLUSTER_NAME)


##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	make helm
	helm install $(HELM_RELEASE_NAME) charts/k8s-controller

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	helm uninstall $(HELM_RELEASE_NAME)

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

PATHPLUSLOCALBIN = PATH=$(LOCALBIN):$$PATH

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
K3D ?= $(LOCALBIN)/k3d
CTLPTL ?= $(LOCALBIN)/ctlptl
HELMIFY ?= $(LOCALBIN)/helmify

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.9.0

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	rm $(KUSTOMIZE)
	curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

# Local environment setup
.PHONY: setup
setup:
	brew bundle
	KUBECTL_INSTALLED=$$(asdf plugin list | grep kubectl || true); \
	if [ -z "$${KUBECTL_INSTALLED}" ]; then \
		asdf plugin add kubectl; \
	fi
	asdf install
.PHONY: start
start: $(CTLPTL)
	$(PATHPLUSLOCALBIN) $(CTLPTL) apply -f cluster.yaml
stop: $(CTLPTL)
	$(CTLPTL) delete -f cluster.yaml

# Helmify stuff

.PHONY: helmify
helmify: $(HELMIFY) ## Download helmify locally if necessary.
$(HELMIFY): $(LOCALBIN)
	test -s $(LOCALBIN)/helmify -- k8s-controller || GOBIN=$(LOCALBIN) go install github.com/arttor/helmify/cmd/helmify@latest

.PHONY: ctlptl
ctlptl: $(CTLPTL) ## Download ctlptl if necessary
$(CTLPTL): $(LOCALBIN)
	test -s $(LOCALBIN)/ctlptl || GOBIN=$(LOCALBIN) go install github.com/tilt-dev/ctlptl/cmd/ctlptl@v0.8.18
