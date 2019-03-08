# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.
FROM golang:1.10.3 as ks-apiserver-builder

COPY / /go/src/kubesphere.io/kubesphere
WORKDIR /go/src/kubesphere.io/kubesphere

RUN go build -o ks-apiserver cmd/ks-apiserver/apiserver.go

FROM alpine:3.7
RUN apk add --update ca-certificates && update-ca-certificates
COPY --from=ks-apiserver-builder /go/src/kubesphere.io/kubesphere/ks-apiserver /usr/local/bin/
CMD ["sh"]
