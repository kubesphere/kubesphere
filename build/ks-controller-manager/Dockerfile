# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.

# Download dependencies
FROM alpine:3.11 as base_os_context

ARG TARGETARCH
ARG TARGETOS
ARG HELM_VERSION=v3.5.2
ARG KUSTOMIZE_VERSION=v4.2.0

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin

WORKDIR /tmp

RUN apk add --no-cache ca-certificates

# Install helm
ADD https://get.helm.sh/helm-${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz /tmp
RUN tar xvzf /tmp/helm-${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz -C /tmp
RUN mv /tmp/${TARGETOS}-${TARGETARCH}/helm ${OUTDIR}/usr/local/bin/

# install kustomize
ADD https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_${TARGETARCH}.tar.gz /tmp
RUN tar xvzf /tmp/kustomize_${KUSTOMIZE_VERSION}_linux_${TARGETARCH}.tar.gz -C /tmp
RUN mv /tmp/kustomize ${OUTDIR}/usr/local/bin/
    

# Build

FROM golang:1.16.3 as build_context

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/

WORKDIR /workspace
ADD . /workspace/

RUN make ks-controller-manager
RUN mv /workspace/bin/cmd/controller-manager ${OUTDIR}/usr/local/bin/

# Final Image

FROM alpine:3.11 

COPY --from=base_os_context /out/ /
COPY --from=build_context /out/ /

WORKDIR /

EXPOSE 8443 8080

CMD ["sh"]