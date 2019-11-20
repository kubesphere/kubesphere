# How to build KubeSphere?

This document walks you through how to get started building KubeSphere in your local environment.

## Preparing the environment

### Go

KubeSphere development is based on [Kubernetes](https://github.com/kubernetes/kubernetes), both of them are written in [Go](http://golang.org/). If you don't have a Go development environment, please [set one up](http://golang.org/doc/code.html).

| Kubernetes     | requires Go |
|----------------|-------------|
| 1.13+          | >= 1.12     |

> Tips:
> - Ensure your GOPATH and PATH have been configured in accordance with the Go
environment instructions.
> - It's recommended to install [macOS GNU tools](https://www.topbug.net/blog/2013/04/14/install-and-use-gnu-command-line-tools-in-mac-os-x) for Mac OS.

### Docker

KubeSphere components are often deployed as containers in Kubernetes. If you need to rebuild the KubeSphere components in the Kubernetes cluster, you will need to [install Docker](https://docs.docker.com/install/).


### Dependency management

KubeSphere uses [Go Modules](https://github.com/golang/go/wiki/Modules) to manage dependencies in the `vendor/` tree.

> Note: Kubesphere uses the `go module` to manage dependencies, but the kubesphere development process still relies on `GOPATH`
> In the CRD development process, you need to use tools to automatically generate code. The tools used by kubesphere still need to rely on `GOPATH`.

> Kubesphere has a large number of Chinese contributors. 
> These contributors may encounter network problems when pulling the go module.We recommend using [goproxy.cn](https://goproxy.cn) as the proxy.

## Building KubeSphere Core on a local OS/shell environment

### For Quick Taste Binary

When you go get kubesphere, you can choose the version you want to get: `go get kubesphere.io/kubesphere@version-you-want`

> Note: Before getting kubesphere, you need to synchronize the contents of the `replace` section of the go.mod file of the kubesphere you want to version.

```bash
mkdir ks-tmp
cd ks-tmp
echo 'module kubesphere' > go.mod
echo 'replace (
        github.com/Sirupsen/logrus v1.4.1 => github.com/sirupsen/logrus v1.4.1
      	github.com/kiali/kiali => github.com/kubesphere/kiali v0.15.1-0.20190407071308-6b5b818211c3
      	github.com/kubernetes-sigs/application => github.com/kubesphere/application v0.0.0-20190518133311-b9d9eb0b5cf7
      )' >> go.mod

GO111MODULE=on go get kubesphere.io/kubesphere@d649e3d0bbc64bfba18816c904819e4850d021e0
GO111MODULE=on go build -o ks-apiserver kubesphere.io/kubesphere/cmd/ks-apiserver # build ks-apiserver
GO111MODULE=on go build -o ks-apigateway kubesphere.io/kubesphere/cmd/ks-apigateway # build ks-apigateway
GO111MODULE=on go build -o ks-controller-manager kubesphere.io/kubesphere/cmd/controller-manager # build ks-controller-manager
GO111MODULE=on go build -o ks-iam kubesphere.io/kubesphere/cmd/ks-iam # build ks-iam
```

### For Building KubeSphere Core Images

KubeSphere components are often deployed as a container in a kubernetes cluster, you may need to build a Docker image locally.

1. Clone repo to local

```bash
git clone https://github.com/kubesphere/kubesphere.git
cd kubesphere
```

2. Run Docker command to build image

```bash
# $REPO is the docker registry to push to
# $Tag is the tag name of the docker image
# The full go build process will be executed in the Dockerfile, so you may need to set GOPROXY in it.
docker build -f build/ks-apigateway/Dockerfile -t $REPO/ks-apigateway:$TAG .
docker build -f build/ks-apiserver/Dockerfile -t $REPO/ks-apiserver:$TAG .
docker build -f build/ks-iam/Dockerfile -t $REPO/ks-account:$TAG .
docker build -f build/ks-controller-manager/Dockerfile -t $REPO/ks-controller-manager:$TAG .
docker build -f ./pkg/db/Dockerfile -t $REPO/ks-devops:flyway-$TAG ./pkg/db/
```

### For KubeSphere Core local development building.


1. Create a `kubesphere` work directory under `GOPATH` and clone the source code
```bash
mkdir -p $GOPATH/src/kubesphere.io/
cd $GOPATH/src/kubesphere.io/
git clone https://github.com/kubesphere/kubesphere
```

2. Use make to build binary
```bash
make ks-apiserver # Build ks-apiserver binary
make ks-iam # Build ks-iam binary
make controller-manager # Build ks-controller-manager binary
make ks-apigateway # Build ks-apigateway binary
```

If you need to build a docker image, you can refer to the previous section.

### Test

In the development process, it is recommended to use local Kubernetes clusters, such as [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/), or to install an single-node [all-in-one](https://github.com/kubesphere/kubesphere#all-in-one) environment (Kubernetes-based) for quick testing.

> Tip: It also supports to use Docker for Desktop ships with Kubernetes as the test environment.


## Building KubeSphere Other Module

Kubesphere has quite a few modules such as ServiceMesh, DevOps, Logging...
Some of these modules have unique build methods, we recommend that you refer to the documentation related to the components.

