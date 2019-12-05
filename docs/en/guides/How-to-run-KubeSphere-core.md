# How to run KubeSphere core in local

This document will explain how to run KubeSphere apiserver locally.

> Modules similar to KubeSphere apiserver are KubeSphere controller-manageer, Kubesphere iam (also known as KubeSphere account), KubeSphere api-gateway. 
> If you need to run these modules locally, you can refer to this document for configuration.

## Prepare: Build KubeSphere Core component

In the document [How-to-build](How-to-build.md) We learned how to build KubeSphere locally. Make sure you could build KubeSphere Core modules accordingly.

## 1. Set up a Kubernetes cluster with KubeSphere installed

KubeSphere relies on some external modules during development, and these modules are already included in the installed KubeSphere.

You can quickly install a KubeSphere cluster by referring to this [documentation] (https://kubesphere.io/en/install).

## 2. Use kubectl to access the development cluster locally

You can refer to this [document](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to install and configure kubectl locally.

## 3. Understand KubeSphere Core's configuration

KubeSphere uses [viper](https://github.com/spf13/viper) to manage the configuration. KubeSphere supports setting up using command line arguments and configuration files.

> We recommend that you use a configuration file for configuration during local development.

KubeSphere apiserver needs to communicate with many modules. When you run Kubesphere, you can choose to configure the seperate modules only you care about.

During the development of KubeSphere apiserver, you must configure at least the relevant part of Kubernetes to ensure that KubeSphere apiserver can be started.

Below is a sample configuration of KubeSphere apiserver:

> Note: In the default configuration, we use Kubernetes service name to access the service. 
> In a remote cluster, you may need to use external network exposure to connect to the cluster's internal services.
> Or you can refer to the [documentation](How-to-connect-remote-service.md) to use `telepresence` to connect to remote services

```yaml
kubernetes:
  kubeconfig: "/Users/kubesphere/.kube/config"
  master: https://192.168.0.8:6443
  qps: 1e+06
  burst: 1000000
ldap:
  host: openldap.kubesphere-system.svc:389
  managerDN: cn=admin,dc=kubesphere,dc=io
  managerPassword: admin
  userSearchBase: ou=Users,dc=kubesphere,dc=io
  groupSearchBase: ou=Groups,dc=kubesphere,dc=io
redis:
  host: redis.kubesphere-system.svc
  port: 6379
  password: ""
  db: 0
s3:
  endpoint: http://minio.kubesphere-system.svc:9000
  region: us-east-1
  disableSSL: true
  forcePathStyle: true
  accessKeyID: openpitrixminioaccesskey
  secretAccessKey: openpitrixminiosecretkey
  bucket: s2i-binaries
mysql:
  host: mysql.kubesphere-system.svc:3306
  username: root
  password: password
  maxIdleConnections: 100
  maxOpenConnections: 100
  maxConnectionLifeTime: 10s

devops:
  host: http://ks-jenkins.kubesphere-devops-system.svc/
  username: admin
  password: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFkbWluQGt1YmVzcGhlcmUuaW8iLCJleHAiOjE4MTYyMzkwMjIsInVzZXJuYW1lIjoiYWRtaW4ifQ.okmNepQvZkBRe1M8z2HAWRN0AVj9ooVu79IafHKCjZI
  maxConnections: 100
sonarQube:
  host: http://192.168.0.8:32297
  token: 4e51de276f1fd0eb3a20b58e523d43ce76347302


openpitrix:
  runtimeManagerEndpoint:    "openpitrix-runtime-manager.openpitrix-system.svc:9103"
  clusterManagerEndpoint:    "openpitrix-cluster-manager.openpitrix-system.svc:9104"
  repoManagerEndpoint:       "openpitrix-repo-manager.openpitrix-system.svc:9101"
  appManagerEndpoint:        "openpitrix-app-manager.openpitrix-system.svc:9102"
  categoryManagerEndpoint:   "openpitrix-category-manager.openpitrix-system.svc:9113"
  attachmentManagerEndpoint: "openpitrix-attachment-manager.openpitrix-system.svc:9122"
  repoIndexerEndpoint:       "openpitrix-repo-indexer.openpitrix-system.svc:9108"

monitoring:
  endpoint: http://prometheus-k8s.kubesphere-monitoring-system.svc:9090
  secondaryEndpoint: http://prometheus-k8s-system.kubesphere-monitoring-system.svc:9090

logging:
  host: http://elasticsearch-logging-data.kubesphere-logging-system.svc.cluster.local:9200
  indexPrefix: ks-logstash-log

alerting:
  endpoint: http://alerting.kubesphere-alerting-system.svc

notification:
  endpoint: http://notification.kubesphere-alerting-system.svc
```

## 4. Set Up KubeSphere Core's configuration

The KubeSphere Core module will read the `kubesphere.yaml` file in the current directory and the `kubesphere.yaml` file in the `/etc/kubesphere` directory, then load the configuration at startup.
You can choose a path to set your configuration locally.

## 5. Run KubeSphere apiserver 

You can execute `go run cmd/ks-apiserver/apiserver.go` in the `$GOPATH/src/kubesphere.io/kubesphere` directory to start KubeSphere apiserver

> If you want to understand the specific meaning of each configuration, you can view it by `go run cmd/ks-apiserver/apiserver.go --help` or read the module's design and developer documentation.
