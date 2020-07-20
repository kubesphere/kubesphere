# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.
FROM alpine:3.11

COPY  /bin/cmd/controller-manager /usr/local/bin/

RUN apk add --no-cache ca-certificates

EXPOSE 8443 8080

CMD ["sh"]