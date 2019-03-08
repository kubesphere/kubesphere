#!/usr/bin/env bash

 docker build -f build/ks-apigateway/Dockerfile -t kubespheredev/ks-apigateway:latest .
 docker build -f build/ks-apiserver/Dockerfile -t kubespheredev/ks-apiserver:latest .