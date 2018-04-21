# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

FROM kubesphere/kubesphere-builder as builder

WORKDIR /go/src/kubesphere.io/kubesphere/
COPY . .

RUN go generate kubesphere.io/kubesphere/pkg/version && \
	go install  kubesphere.io/kubesphere/cmd/...

FROM alpine:3.6
RUN apk add --update ca-certificates && update-ca-certificates
COPY --from=builder /go/bin/* /usr/local/bin/

CMD ["sh"]
