#!/usr/bin/env bash

 docker build -f build/ks-apigateway/Dockerfile -t kubespheredev/ks-apigateway:latest .
 docker build -f build/ks-apiserver/Dockerfile -t kubespheredev/ks-apiserver:latest .
 docker build -f build/ks-iam/Dockerfile -t kubespheredev/ks-account:latest .

 docker build -f build/controller-manager/Dockerfile -t kubespheredev/ks-controller-manager:latest .

 docker build -f ./pkg/db/Dockerfile -t kubespheredev/ks-devops:flyway ./pkg/db/
