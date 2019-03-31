#!/usr/bin/env bash

 docker build -f build/ks-apigateway/Dockerfile -t kubespheredev/ks-apigateway:latest .
 docker build -f build/ks-apiserver/Dockerfile -t kubespheredev/ks-apiserver:latest .
 docker build -f build/ks-iam/Dockerfile -t kubespheredev/ks-iam:latest .

 docker build -f build/controller-manager/Dockerfile -t kubespheredev/ks-controller-manager:latest .
