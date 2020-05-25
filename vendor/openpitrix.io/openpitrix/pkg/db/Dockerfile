# Copyright 2017 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

FROM openpitrix/flyway:alpine

USER root
RUN apk add --no-cache mysql-client

COPY ./schema /flyway/sql
COPY ./ddl /flyway/sql/ddl
COPY ./scripts /flyway/sql/ddl
