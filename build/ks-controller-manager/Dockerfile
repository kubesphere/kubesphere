# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.
FROM alpine/helm:3.4.2 as helm-base

FROM alpine:3.11

RUN apk add --no-cache ca-certificates
COPY --from=helm-base /usr/bin/helm /usr/bin/helm
COPY  /bin/cmd/controller-manager /usr/local/bin/

EXPOSE 8443 8080

CMD ["sh"]