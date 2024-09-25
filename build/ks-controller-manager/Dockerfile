# Download dependencies
FROM alpine:3.18.3 AS base_os_context

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/
RUN apk add --no-cache ca-certificates curl

ARG TARGETARCH

ENV ARCH=$TARGETARCH

RUN curl -LO https://github.com/kubesphere/telemetry/releases/download/v1.0.0/telemetry_1.0.0_linux_${ARCH}.tar.gz && \
    tar -zxvf telemetry_1.0.0_linux_${ARCH}.tar.gz -C ${OUTDIR}/usr/local/bin/ telemetry

# Add KS Core Helm Chart
COPY config/ks-core ${OUTDIR}/var/helm-charts/ks-core

# Build
FROM golang:1.21.5 AS build_context

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/

WORKDIR /workspace
ADD . /workspace/

RUN make ks-controller-manager
RUN mv /workspace/bin/cmd/controller-manager ${OUTDIR}/usr/local/bin/

# Final Image
FROM alpine:3.18.3

COPY --from=base_os_context /out/ /
COPY --from=build_context /out/ /

WORKDIR /

EXPOSE 8443
CMD ["sh"]
