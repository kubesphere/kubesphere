# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.

# Build
FROM golang:1.19.4 as build_context

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/

WORKDIR /workspace
ADD . /workspace/

RUN make ks-apiserver
RUN mv /workspace/bin/cmd/ks-apiserver ${OUTDIR}/usr/local/bin/

##############
# Final image
#############

FROM alpine:3.18.3

COPY --from=build_context /out/ /

WORKDIR /

EXPOSE 9090
CMD ["sh"]
