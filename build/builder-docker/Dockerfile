# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

FROM golang:1.10.2-alpine3.7 as builder

RUN apk add --no-cache git curl openssl

RUN go get github.com/tools/godep
#RUN go get github.com/emicklei/go-restful
#RUN go get github.com/golang/glog
#RUN go get github.com/spf13/pflag
RUN go get golang.org/x/tools/cmd/goimports

FROM golang:1.10.2-alpine3.7

RUN apk add --no-cache git make curl openssl jq rsync godep

COPY --from=builder /go/bin /go/bin
