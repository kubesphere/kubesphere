# Developing for KubeSphere

The [community repository](https://github.com/kubesphere) hosts all information about
building KubeSphere from source, how to contribute code and documentation, who to contact about what, etc. If you find a requirement that this doc does not capture, or if you find other docs with references to requirements that are not simply links to this doc, please [submit an issue](https://github.com/kubesphere/kubesphere/issues/new).

----

## To start developing KubeSphere

First of all, you should fork the project. Then follow one of the three options below to develop the project. Please note you should replace the official repo when using __go get__ or __git clone__ below with your own one.

### 1. You have a working [Docker Compose](https://docs.docker.com/compose/install) environment [recommend].
>You need to install [Docker](https://docs.docker.com/engine/installation/) first.

```shell
$ git clone https://github.com/kubesphere/kubesphere
$ cd kubesphere
$ make build
$ make compose-up
```

Exit docker runtime environment
```shell
$ make compose-down
```

### 2. You have a working [Docker](https://docs.docker.com/engine/installation/) environment.


Exit docker runtime environment
```shell
$ docker stop $(docker ps -f name=kubesphere -q)
```

### 3. You have a working [Go](prereqs.md#setting-up-go) environment.

- Install [protoc compiler](https://github.com/google/protobuf/releases/)
- Install protoc plugin:

```shell
$ go get github.com/golang/protobuf/protoc-gen-go
$ go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
$ go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
$ go get github.com/mwitkow/go-proto-validators/protoc-gen-govalidators
```

- Get kubesphere source code and build service:

```shell
$ go get -d kubesphere.io/kubesphere
$ cd $GOPATH/src/kubesphere.io/kubesphere
$ make generate
$ GOBIN=`pwd`/bin go install ./cmd/...
```

- Start KubeSphere service:


- Exit go runtime environment
```shell
$ ps aux | grep kubesphere- | grep -v grep | awk '{print $2}' | xargs kill -9
```

----

## Test KubeSphere

Visit http://127.0.0.1:9100/swagger-ui in browser, and try it online, or test kubesphere api service via command line:

----

## DevOps

Please check [How to set up DevOps environment](devops.md).
