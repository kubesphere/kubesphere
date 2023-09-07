# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.

# Download dependencies
FROM alpine:3.18.3 as base_os_context

ARG INGRESS_NGINX_VERSION=4.2.5

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/var/helm-charts

WORKDIR /tmp

RUN apk add --no-cache ca-certificates


# Install Nginx Ingress Helm Chart
ADD https://github.com/kubernetes/ingress-nginx/releases/download/helm-chart-${INGRESS_NGINX_VERSION}/ingress-nginx-${INGRESS_NGINX_VERSION}.tgz /tmp
RUN tar xvzf /tmp/ingress-nginx-${INGRESS_NGINX_VERSION}.tgz -C /tmp
RUN mv /tmp/ingress-nginx ${OUTDIR}/var/helm-charts/

# Build

FROM golang:1.19.4 as build_context

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/
RUN mkdir -p ${OUTDIR}/var/helm-charts

WORKDIR /workspace
ADD . /workspace/

RUN make ks-controller-manager
RUN mv /workspace/bin/cmd/controller-manager ${OUTDIR}/usr/local/bin/

# Copy gateway config and helm chart
RUN mv /workspace/config/gateway ${OUTDIR}/var/helm-charts/
RUN mv /workspace/config/watches.yaml ${OUTDIR}/var/helm-charts/

# Final Image

FROM alpine:3.18.3

COPY --from=base_os_context /out/ /
COPY --from=build_context /out/ /

WORKDIR /

EXPOSE 8443 8080

CMD ["sh"]
