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
