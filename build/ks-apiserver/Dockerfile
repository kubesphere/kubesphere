# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.
# Build
FROM golang:1.18.4 as builder

WORKDIR /src
ADD . /src/

RUN make ks-apiserver

##############
# Final image
#############

FROM alpine:3.16

COPY --from=builder /src/bin/cmd/ks-apiserver /usr/local/bin/ks-apiserver

WORKDIR /

EXPOSE 9090
CMD ["sh"]
