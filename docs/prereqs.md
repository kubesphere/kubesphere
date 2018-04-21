# Developing for KubeSphere [deprecated]

This document is intended to be the canonical source of truth for things like
supported toolchain versions for building KubeSphere.
If you find a requirement that this doc does not capture, or if you find other
docs with references to requirements that are not simply links to this doc,
please [submit an issue](https://github.com/kubesphere/kubesphere/issues/new).

This document is intended to be relative to the branch in which it is found.
It is guaranteed that requirements will change over time for the development
branch, but release branches should not change.

- [Prerequisites](#prerequisites)
  - [Setting up Go](#setting-up-go)
  - [Setting up Swagger](#setting-up-swagger)
- [To start developing KubeSphere](#to-start-developing-kubesphere)
- [DevOps](#devops)

## Prerequisites

KubeSphere only has few external dependencies you need to setup before being
able to build and run the code.

### Setting up Go

KubeSphere written in the [Go](http://golang.org) programming language.
To build, you'll need a Go (version 1.9+) development environment.
If you haven't set up a Go development environment, please follow
[these instructions](https://golang.org/doc/install)
to install the Go tools.

Set up your GOPATH and add a path entry for Go binaries to your PATH. Typically
added to your ~/.profile:

```shell
$ export GOPATH=~/go
$ export PATH=$PATH:$GOPATH/bin
```

### Setting up Swagger

KubeSphere is using [OpenAPI/Swagger](https://swagger.io) to develop API, so follow
[the instructions](https://github.com/go-swagger/go-swagger/tree/master/docs) to
install Swagger. If you are not familar with Swagger, please read the
[tutorial](http://apihandyman.io/writing-openapi-swagger-specification-tutorial-part-1-introduction/#writing-openapi-fka-swagger-specification-tutorial). If you install Swagger using docker distribution,
please run

```shell
$ docker pull quay.io/goswagger/swagger
$ alias swagger="docker run --rm -it -e GOPATH=$GOPATH:/go -v $HOME:$HOME -w $(pwd) quay.io/goswagger/swagger"
$ swagger version
```

## To start developing KubeSphere

There are two options to get KubeSphere source code and build the project:

**You have a working Go environment.**

```shell
$ go get -d kubesphere.io/kubesphere
$ cd $GOPATH/src/kubesphere.io/kubesphere
$ make all
```

**You have a working Docker environment.**

```shell
$ git clone https://github.com/kubesphere/kubesphere
$ cd kubesphere
$ make
```

## DevOps

Please check [How to set up DevOps environment](devops.md)
