#!/bin/bash

set -e

usage() {
    cat <<EOF
Generate certificate suitable for use with an sidecar-injector webhook service.
This script uses k8s' CertificateSigningRequest API to a generate a
certificate signed by k8s CA suitable for use with sidecar-injector webhook
services. This requires permissions to create and approve CSR. See
https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster for
detailed explantion and additional instructions.
The server key/cert k8s CA cert are stored in a k8s secret.
usage: ${0} [OPTIONS]
The following flags are required.
       --service          Service name of webhook.
       --namespace        Namespace where webhook service and secret reside.
EOF
    exit 1
}

while [[ $# -gt 0 ]]; do
    case ${1} in
        --service)
            service="$2"
            shift
            ;;
        --namespace)
            namespace="$2"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done

[ -z "${service}" ] && service=webhook-service
[ -z "${namespace}" ] && namespace=default

if [ ! -x "$(command -v openssl)" ]; then
    echo "openssl not found"
    exit 1
fi

# csrName=${service}.${namespace}
CERTSDIR="config/certs"

if [ ! -d ${CERTSDIR} ]; then
  mkdir ${CERTSDIR}
fi

echo "creating certs in certsdir ${CERTSDIR} "

# create cakey
openssl genrsa -out ${CERTSDIR}/ca.key 2048

# create ca.crt
openssl req -x509 -new -nodes -key ${CERTSDIR}/ca.key -subj "/C=CN/ST=HB/O=QC/CN=${service}" -sha256 -days 10000 -out ${CERTSDIR}/ca.crt

# create server.key
openssl genrsa -out ${CERTSDIR}/server.key 2048

# create server.crt
openssl req -new -sha256 -key ${CERTSDIR}/server.key -subj "/C=CN/ST=HB/O=QC/CN=${service}.${namespace}.svc" -out ${CERTSDIR}/server.csr
openssl x509 -req -in ${CERTSDIR}/server.csr -CA ${CERTSDIR}/ca.crt -CAkey ${CERTSDIR}/ca.key -CAcreateserial -out ${CERTSDIR}/server.crt -days 10000 -sha256
