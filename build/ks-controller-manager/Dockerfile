# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.

FROM golang:1.12 as controller-manager-builder

COPY / /go/src/kubesphere.io/kubesphere
WORKDIR /go/src/kubesphere.io/kubesphere

RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 GOFLAGS=-mod=vendor go build --ldflags "-extldflags -static" -o controller-manager ./cmd/controller-manager/

FROM alpine:3.7
RUN apk add --update ca-certificates && update-ca-certificates
COPY --from=controller-manager-builder /go/src/kubesphere.io/kubesphere/controller-manager /usr/local/bin/
CMD controller-manager
