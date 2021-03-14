# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.
FROM alpine:3.11

ARG HELM_VERSION=v3.5.2

RUN apk add --no-cache ca-certificates
# install helm
RUN wget https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    tar xvf helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    rm helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/bin/ && \
    rm -rf linux-amd64
# To speed up building process, we copy binary directly from make
# result instead of building it again, so make sure you run the 
# following command first before building docker image
#   make ks-apiserver
#
COPY  /bin/cmd/ks-apiserver /usr/local/bin/

EXPOSE 9090
CMD ["sh"]
