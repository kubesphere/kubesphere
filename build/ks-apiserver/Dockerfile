# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.
FROM alpine:3.9
RUN apk add --update ca-certificates && update-ca-certificates
COPY  /bin/cmd/ks-apiserver /usr/local/bin/
CMD ["sh"]
