# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.

FROM golang:1.12 as ks-apiserver-builder

COPY / /go/src/kubesphere.io/kubesphere

WORKDIR /go/src/kubesphere.io/kubesphere
RUN GIT_VERSION=$(git describe --always --dirty) && \
    GIT_HASH=$(git rev-parse HEAD) && \
    BUILDDATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') && \
    CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 GOFLAGS=-mod=vendor \
    go build -i -ldflags \
    '-w -s -X kubesphere.io/kubesphere/pkg/version.version=$(GIT_VERSION) \
     -X kubesphere.io/kubesphere/pkg/version.gitCommit=$(GIT_HASH) \
     -X kubesphere.io/kubesphere/pkg/version.buildDate=$(BUILDDATE)' \
    -o ks-apiserver cmd/ks-apiserver/apiserver.go

FROM alpine:3.9
RUN apk add --update ca-certificates && update-ca-certificates
COPY --from=ks-apiserver-builder /go/src/kubesphere.io/kubesphere/ks-apiserver /usr/local/bin/
CMD ["sh"]
