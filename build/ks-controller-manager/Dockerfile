# Download dependencies
FROM alpine:3.20.3 AS base_os_context

RUN apk add --no-cache ca-certificates curl

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/

ARG TELEMETRY_VERSION=1.0.1
ARG TARGETARCH
RUN curl -LO https://github.com/kubesphere/telemetry/releases/download/v${TELEMETRY_VERSION}/telemetry_${TELEMETRY_VERSION}_linux_${TARGETARCH}.tar.gz && \
    tar -zxvf telemetry_${TELEMETRY_VERSION}_linux_${TARGETARCH}.tar.gz -C ${OUTDIR}/usr/local/bin/ telemetry

# Add KS Core Helm Chart
COPY config/ks-core ${OUTDIR}/var/helm-charts/ks-core

# Build
FROM golang:1.22.8 AS build_context

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/

WORKDIR /workspace
ADD . /workspace/

RUN make ks-controller-manager
RUN mv /workspace/bin/cmd/controller-manager ${OUTDIR}/usr/local/bin/

# Final Image
FROM alpine:3.20.3

COPY --from=base_os_context /out/ /
COPY --from=build_context /out/ /

WORKDIR /

EXPOSE 8443

CMD ["sh"]