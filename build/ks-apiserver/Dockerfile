# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.
FROM alpine:3.9

COPY  /bin/cmd/ks-apiserver /usr/local/bin/

RUN apk add --update ca-certificates && \
    update-ca-certificates && \
    adduser -D -g kubesphere -u 1002 kubesphere && \
    chown -R kubesphere:kubesphere /usr/local/bin/ks-apiserver

USER kubesphere
CMD ["sh"]
