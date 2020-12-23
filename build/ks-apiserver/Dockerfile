# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.
FROM alpine:3.11

ENV BASE_URL=https://get.helm.sh
ENV TAR_FILE=helm-v3.4.2-linux-amd64.tar.gz
RUN apk add --update --no-cache curl ca-certificates && curl -L ${BASE_URL}/${TAR_FILE} |tar xvz &&  mv linux-amd64/helm /usr/bin/helm && chmod +x /usr/bin/helm && rm -rf linux-amd64 && apk del curl && rm -f /var/cache/apk/*

# To speed up building process, we copy binary directly from make
# result instead of building it again, so make sure you run the 
# following command first before building docker image
#   make ks-apiserver
#
COPY  /bin/cmd/ks-apiserver /usr/local/bin/

EXPOSE 9090
CMD ["sh"]
