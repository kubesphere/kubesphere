# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

FROM golang:1.10.3 as ks-apigateway-builder

COPY / /go/src/kubesphere.io/kubesphere
RUN git clone --single-branch -b v0.11.4 -q https://github.com/mholt/caddy /go/src/github.com/mholt/caddy
WORKDIR /go/src/github.com/mholt/caddy
RUN sed -i "/\/\/ This is where other plugins get plugged in (imported)/a\\\t\
    _ \"kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/authenticate\"\n\t\
    _ \"kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/authentication\"\n\t\
    _ \"kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/swagger\""\
    caddy/caddymain/run.go && \
    sed -i "/\/\/ github.com\/BTBurke\/caddy-jwt/a\\\t\"authenticate\",\n\t\"authentication\",\n\t\"swagger\","\
    caddyhttp/httpserver/plugin.go && \
    go install ./caddy && \
    go run /go/src/kubesphere.io/kubesphere/tools/cmd/doc-gen/main.go --output=/go/src/kubesphere.io/kubesphere/install/swagger-ui/api.json

FROM alpine:3.7
RUN apk add --update ca-certificates && update-ca-certificates
COPY --from=ks-apigateway-builder /go/bin/* /usr/local/bin/
COPY --from=ks-apigateway-builder /go/src/kubesphere.io/kubesphere/install/swagger-ui /var/static/swagger-ui
CMD ["sh"]