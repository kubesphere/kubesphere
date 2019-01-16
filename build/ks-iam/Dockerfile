# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.
FROM golang:1.10.3 as ks-iam-builder

COPY / /go/src/kubesphere.io/kubesphere
WORKDIR /go/src/kubesphere.io/kubesphere

RUN go build -ldflags \
    "-X kubesphere.io/kubesphere/pkg/version.Version=`git describe --tags` \
    -X kubesphere.io/kubesphere/pkg/version.BuildTime=`date +%FT%T%z`" \
    -o ks-iam cmd/ks-iam/main.go

FROM alpine:3.7
RUN apk add --update ca-certificates && update-ca-certificates
COPY --from=ks-iam-builder /go/src/kubesphere.io/kubesphere/ks-iam /usr/local/bin/
CMD ["sh"]
