# Copyright 2020 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by an Apache license
# that can be found in the LICENSE file.
FROM alpine:3.11

RUN apk add --no-cache ca-certificates 

# To speed up building process, we copy binary directly from make
# result instead of building it again, so make sure you run the 
# following command first before building docker image
#   make ks-apiserver
#
COPY  /bin/cmd/ks-apiserver /usr/local/bin/

EXPOSE 9090
CMD ["sh"]
