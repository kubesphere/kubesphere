FROM ubuntu:18.04 as binarydownloader
ARG HELM_VERSION=v3.5.2
ARG KUSTOMIZE_VERSION=v4.0.5
ARG KUBE_VERSION=v1.16.4
ARG ETCD_VERSION=v3.3.11
RUN apt update && apt -y install wget curl
ADD ./hack /opt/hack/

RUN bash /opt/hack/build-scripts/get-helm.sh
RUN bash /opt/hack/build-scripts/get-kustomize.sh
RUN bash /opt/hack/build-scripts/get-etcd.sh
RUN bash /opt/hack/build-scripts/get-kubernetes.sh


FROM golang:1.13 as builder
ADD . /opt/kubesphere/
WORKDIR /opt/kubesphere/
RUN apt update && apt -y install sudo
RUN bash /opt/kubesphere/hack/install_kubebuilder.sh
COPY --from=binarydownloader /usr/local/kubernetes/server/bin/kube-apiserver /usr/local/kubebuilder/bin/kube-apiserver
COPY --from=binarydownloader /usr/local/kubernetes/server/bin/kubectl /usr/local/kubebuilder/bin/kubectl
COPY --from=binarydownloader /usr/local/bin/etcd /usr/local/kubebuilder/bin/etcd
RUN bash -c "source /opt/kubesphere/hack/init_env.sh && make ks-apiserver"
RUN bash -c "source /opt/kubesphere/hack/init_env.sh && make ks-controller-manager"


FROM alpine:3.11 as ks-apiserver
RUN apk add --no-cache ca-certificates
COPY --from=binarydownloader /usr/bin/helm /usr/bin/helm
COPY --from=builder /opt/kubesphere/bin/cmd/ks-apiserver /usr/local/bin/

EXPOSE 9090

CMD ["sh"]

FROM alpine:3.11 as ks-controller-manager
RUN apk add --no-cache ca-certificates
COPY --from=binarydownloader /usr/bin/helm /usr/bin/helm
COPY --from=binarydownloader /usr/bin/kustomize /usr/bin/kustomize
COPY --from=builder /opt/kubesphere/bin/cmd/controller-manager /usr/local/bin/

EXPOSE 8443 8080

CMD ["sh"]