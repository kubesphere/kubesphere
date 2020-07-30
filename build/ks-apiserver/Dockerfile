# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.
FROM alpine:3.11
ENV USER=kubesphere
ENV UID=1002
# docker group
ENV GID=998

RUN apk add --no-cache ca-certificates && \
    addgroup --gid "$GID" "$USER" && \
    adduser -D -g "" --ingroup "$USER" -u "$UID" "$USER"

COPY  /bin/cmd/ks-apiserver /usr/local/bin/

RUN chown -R kubesphere:"$GID" /usr/local/bin/ks-apiserver

EXPOSE 9090

USER kubesphere
CMD ["sh"]