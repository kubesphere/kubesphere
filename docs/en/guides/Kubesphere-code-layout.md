# KubeSphere File Tree

This document describes the directory structure of the KubeSphere repository.


```
├── api // Automatically generated API documentation
│   ├── api-rules
│   ├── ks-openapi-spec // REST API documentation provided by kubesphere apiserver
│   └── openapi-spec // REST API documentation provided by kubesphere apiserver
├── build // Dockerfile 
│   ├── hypersphere 
│   ├── ks-apigateway
│   ├── ks-apiserver
│   ├── ks-controller-manager
│   ├── ks-iam
│   └── ks-network
├── cmd // Main applications for KubeSphere.
│   ├── controller-manager  // Kubesphere Controller Manger, used to reconcile KubeSphere CCRD
│   │   └── app
│   ├── hypersphere 
│   ├── ks-apigateway // KubeSphere API gateway
│   │   └── app
│   ├── ks-apiserver // KubeSphere REST API server
│   │   └── app
│   ├── ks-iam // KubeSphere iam service
│   │   └── app
│   └── ks-network
├── config // CRD config files
│   ├── crds // CRD yaml files
│   ├── default // kustomization yaml files
│   ├── manager // controller manager yaml files
│   ├── rbac // rbac yaml files
│   ├── samples // CRD sample
│   └── webhook // webhppk yaml files
├── docs 
│   ├── en
│   │   ├── concepts-and-designs
│   │   └── guides
│   └── images
├── hack // Script files to help people develop
│   └── lib
├── pkg // Library code. 
│   ├── api // Structure definitions for REST APIs
│   │   ├── devops
│   │   ├── logging
│   │   └── monitoring
│   ├── apigateway 
│   │   └── caddy-plugin
│   ├── apis // Structure definitions for CRDs
│   │   ├── devops
│   │   ├── network
│   │   ├── servicemesh
│   │   └── tenant
│   ├── apiserver // REST API parameter processing
│   │   ├── components
│   │   ├── devops
│   │   ├── git
│   │   ├── iam
│   │   ├── logging
│   │   ├── monitoring
│   │   ├── openpitrix
│   │   ├── operations
│   │   ├── quotas
│   │   ├── registries
│   │   ├── resources
│   │   ├── revisions
│   │   ├── routers
│   │   ├── runtime
│   │   ├── servicemesh
│   │   ├── tenant
│   │   ├── terminal
│   │   ├── workloadstatuses
│   │   └── workspaces
│   ├── client //Automatically generated CRD client
│   │   ├── clientset
│   │   ├── informers
│   │   └── listers
│   ├── constants // common constants
│   ├── controller // controller manger's reconciliation logic
│   │   ├── application
│   │   ├── clusterrolebinding
│   │   ├── destinationrule
│   │   ├── job
│   │   ├── namespace
│   │   ├── network
│   │   ├── s2ibinary
│   │   ├── s2irun
│   │   ├── storage
│   │   ├── virtualservice
│   │   └── workspace
│   ├── db // Database ORM Framework
│   │   ├── ddl
│   │   ├── schema
│   │   └── scripts
│   ├── gojenkins // Jenkins Go Client
│   │   ├── _tests
│   │   └── utils
│   ├── informers
│   ├── kapis // REST API registration
│   │   ├── devops
│   │   ├── iam
│   │   ├── logging
│   │   ├── monitoring
│   │   ├── openpitrix
│   │   ├── operations
│   │   ├── resources
│   │   ├── servicemesh
│   │   ├── tenant
│   │   └── terminal
│   ├── models // Data processing part of REST API
│   │   ├── components
│   │   ├── devops
│   │   ├── git
│   │   ├── iam
│   │   ├── kubeconfig
│   │   ├── kubectl
│   │   ├── log
│   │   ├── metrics
│   │   ├── nodes
│   │   ├── openpitrix
│   │   ├── quotas
│   │   ├── registries
│   │   ├── resources
│   │   ├── revisions
│   │   ├── routers
│   │   ├── servicemesh
│   │   ├── status
│   │   ├── storage
│   │   ├── tenant
│   │   ├── terminal
│   │   ├── workloads
│   │   └── workspaces
│   ├── server // Data processing part of REST API
│   │   ├── config
│   │   ├── errors
│   │   ├── filter
│   │   ├── options
│   │   └── params
│   ├── simple // common clients
│   │   └── client
│   ├── test
│   ├── utils // common utils
│   │   ├── hashutil
│   │   ├── idutils
│   │   ├── iputil
│   │   ├── jsonutil
│   │   ├── jwtutil
│   │   ├── k8sutil
│   │   ├── net
│   │   ├── readerutils
│   │   ├── reflectutils
│   │   ├── signals
│   │   ├── sliceutil
│   │   ├── stringutils
│   │   └── term
│   ├── version
│   └── webhook
├── test // e2e test code
│   ├── e2e
├── tools // tools to genereate API docs
│   ├── cmd
│   │   ├── crd-doc-gen // gen CRD API docs
│   │   └── doc-gen // gen REST API docs
│   └── lib

```
