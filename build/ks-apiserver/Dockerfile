# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.
FROM golang:1.12 as ks-apiserver-builder

COPY / /go/src/kubesphere.io/kubesphere

WORKDIR /go/src/kubesphere.io/kubesphere
RUN CGO_ENABLED=0 GO111MODULE=off GOOS=linux GOARCH=amd64 go build -i -ldflags '-w -s' -o ks-apiserver cmd/ks-apiserver/apiserver.go

FROM alpine:3.9
RUN apk add --update ca-certificates && update-ca-certificates
COPY --from=ks-apiserver-builder /go/src/kubesphere.io/kubesphere/ks-apiserver /usr/local/bin/
CMD ["sh"]
