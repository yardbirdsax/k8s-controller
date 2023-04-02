# controller
This repository contains a simple Kubernetes controller implementation to learn how developing controllers and CRDs works.

## Tenets

* Build controllers using the pattern described by [the Kubebuilder
  book](https://book.kubebuilder.io/) and the accompanying framework.
* Generate Helm charts from the manifests created by Kubebuilder using
  [`helmify`](github.com/arttor/helmify/cmd/helmify).
* Write end-to-end tests that deploy to a real Kubernetes cluster, similar to what [this
  talk](https://www.youtube.com/watch?v=T4EB0KB1-fc) by Christie Wilson describes.
* Simplify the local development environment setup, including a fully functional Kubernetes cluster
  (used for the aforementioned end-to-end testing) using `ctlptl` by [tilt.dev](https://tilt.dev)
  and [`k3d`](https://github.com/rancher/k3d) for creating a cluster in Docker.

## Description

Right now, there are the following controllers.

### Service Controller

This controller watches for services with the annotation `feiermanfamily.com/labelService: true` and
then attaches a label of `controllerTagged: true` to them. The idea was to learn how to watch for
and modify objects that aren't custom or created by the controller.

## Getting Started

To install tools required for local development, run `make setup`. (Sorry to those of you who are
not using Macs.)

## Running End-To-End Tests

Run `make test-e2e` to provision a local `k3d` cluster, install the controllers via Helm, and run an
end-to-end test.

## Modifying the API definitions
If you are editing the API definitions, generate the manifests and Helm charts such as CRs or CRDs using:

```sh
make helm
```

## Building and pushing Docker images

To build and push Docker images for the controller manager, use the `make docker-build` and `make
docker-push` commands. You can optionally specify a different image name using the `IMG` argument
like this: `make docker-build IMG=ghcr.io/yardbirdsax/k8s-controller/controller`.

## Deploying the controller

You can deploy the controller using the `make deploy` command, and uninstall it using the `make
undeploy` command. If you previously specified a different image name using the `IMG` argument make
sure you do the same here. The `helm` tool deploys the controller using the generated chart.

**NOTE:** Run `make help` for more information on all potential `make` targets

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

