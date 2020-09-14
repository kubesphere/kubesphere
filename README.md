# KubeSphere Container Platform

[![License](http://img.shields.io/badge/license-apache%20v2-blue.svg)](https://github.com/KubeSphere/KubeSphere/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/kubesphere/kubesphere.svg?branch=master)](https://travis-ci.org/kubesphere/kubesphere)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubesphere/kubesphere)](https://goreportcard.com/report/github.com/kubesphere/kubesphere)
[![KubeSphere release](https://img.shields.io/github/release/kubesphere/kubesphere.svg?color=release&label=release&logo=release&logoColor=release)](https://github.com/kubesphere/kubesphere/releases/tag/v3.0.0)

![logo](docs/images/kubesphere-logo.png)

----

## What is KubeSphere

> English | [中文](README_zh.md)

[KubeSphere](https://kubesphere.io/) is a **distributed operating system providing cloud native stack** with [Kubernetes](https://kubernetes.io) as its kernel, and aims to be plug-and-play architecture for third-party applications seamless integration to boost its ecosystem. KubeSphere is also a multi-tenant enterprise-grade container platform with full-stack automated IT operation and streamlined DevOps workflows. It provides developer-friendly wizard web UI, helping enterprises to build out a more robust and feature-rich platform, which includes most common functionalities needed for enterprise Kubernetes strategy, see [Feature List](#features) for details.

The following screenshots give a close insight into KubeSphere. Please check [What is KubeSphere](https://kubesphere.io/docs/introduction/what-is-kubesphere/) for further information.

<table>
  <tr>
      <td width="50%" align="center"><b>Workbench</b></td>
      <td width="50%" align="center"><b>Project Resources</b></td>
  </tr>
  <tr>
     <td><img src="docs/images/console.png"/></td>
     <td><img src="docs/images/project.png"/></td>
  </tr>
  <tr>
      <td width="50%" align="center"><b>CI/CD Pipeline</b></td>
      <td width="50%" align="center"><b>App Store</b></td>
  </tr>
  <tr>
     <td><img src="docs/images/cicd.png"/></td>
     <td><img src="docs/images/app-store.png"/></td>
  </tr>
</table>

## Demo Environment

Using the account `demo1 / Demo123` to log in the [demo environment](https://demo.kubesphere.io/). Please note the account is granted view access. You can also have a quick view of [KubeSphere Demo Video](https://youtu.be/u5lQvhi_Xlc).

## Architecture

KubeSphere uses a loosely-coupled architecture that separates the [frontend](https://github.com/kubesphere/console) from the [backend](https://github.com/kubesphere/kubesphere). External systems can access the components of the backend which are delivered as Docker containers through the REST APIs. See [Architecture](https://kubesphere.io/docs/introduction/architecture/) for details.

![Architecture](docs/images/architecture.png)

## Features

|Feature|Description|
|---|---|
| Provisioning Kubernetes Cluster|Support deploy Kubernetes on your infrastructure out of box, including online and air gapped installation|
| Multi-cluster Management | Provide a centralized control plane to manage multiple Kubernetes Clusters, support application distribution across multiple clusters and cloud providers|
| Kubernetes Resource Management | Provide web console for creating and managing Kubernetes resources, with powerful observability including monitoring, logging, events, alerting and notification |
| DevOps System | Provide out-of-box CI/CD based on Jenkins, and offers automated workflow tools including binary-to-image (B2I) and source-to-image (S2I) |
| Application Store | Provide application store for Helm-based applications, and offers application lifecycle management |
| Service Mesh (Istio-based) | Provide fine-grained traffic management, observability and tracing for distributed microservice applications, provides visualization for traffic topology |
| Rich Observability | Provide multi-dimensional monitoring metrics, and provides multi-tenant logging, events and [auditing](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/) management, support alerting and notification for both application and infrastructure |
| Multi-tenant Management | Provide unified authentication with fine-grained roles and three-tier authorization system, supports AD/LDAP authentication |
| Infrastructure Management | Support node management and monitoring, and supports adding new nodes for Kubernetes cluster |
| Storage Support | Support GlusterFS, CephRBD, NFS, LocalPV (default), etc. open source storage solutions, provides CSI plugins to consume storage from cloud providers |
| Network Support | Support Calico, Flannel, etc., provides [Network Policy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) management, and load balancer plug-in [Porter](https://github.com/kubesphere/porter) for bare metal.|
| GPU Support | Support add GPU node, support vGPU, enables running ML applications on Kubernetes, e.g. TensorFlow |

Please see the [Feature and Benefits](https://kubesphere.io/docs/introduction/features/) for further information.

----

## Latest Release

KubeSphere 3.0.0 is now generally available! See the [Release Notes For 3.0.0](https://kubesphere.io/docs/release/release-v300/) for the updates.

## Installation

KubeSphere can run anywhere from on-premise datacenter to any cloud to edge. In addition, it can be deployed on any version-compatible running Kubernetes cluster.

### QuickStarts

[Quickstarts](https://kubesphere.io/docs/quick-start/) include six hands-on lab exercises that help you quickly get started with KubeSphere.

### Installing on Existing Kubernetes Cluster

- [Installing KubeSphere on Amazone EKS](https://v3-0.docs.kubesphere.io/docs/installing-on-kubernetes/hosted-kubernetes/install-kubesphere-on-eks/)
- [Installing KubeSphere on Azure AKS](https://v3-0.docs.kubesphere.io/docs/installing-on-kubernetes/hosted-kubernetes/install-kubesphere-on-aks/)
- [Installing KubeSphere on Google GKE](https://v3-0.docs.kubesphere.io/docs/installing-on-kubernetes/hosted-kubernetes/install-kubesphere-on-aks/)
- [Installing KubeSphere on DigitalOcean Kubernetes](https://v3-0.docs.kubesphere.io/docs/installing-on-kubernetes/hosted-kubernetes/install-kubesphere-on-do/)
- [Installing KubeSphere on Oracle OKE](https://v3-0.docs.kubesphere.io/docs/installing-on-kubernetes/hosted-kubernetes/install-kubesphere-on-oke/)
- [Installing KubeSphere on Tencent TKE](https://v3-0.docs.kubesphere.io/docs/installing-on-kubernetes/hosted-kubernetes/install-ks-on-tencent-tke/)
- [Installing KubeSphere on Huaweicloud CCE](https://v3-0.docs.kubesphere.io/docs/installing-on-kubernetes/hosted-kubernetes/install-ks-on-huawei-cce/)

### Installing on Linux

- [Installing KubeSphere on Azure VM](https://https://v3-0.docs.kubesphere.io/docs/installing-on-linux/public-cloud/install-ks-on-azure-vms/)
- [Installing KubeSphere on VMware vSphere](https://https://v3-0.docs.kubesphere.io/docs/installing-on-linux/on-premises/install-kubesphere-on-vmware-vsphere/)
- [Installing KubeSphere on QingCloud Instance](https://https://v3-0.docs.kubesphere.io/docs/installing-on-linux/public-cloud/kubesphere-on-qingcloud-instance/)
- [Installing on Alibaba Cloud ECS](https://https://v3-0.docs.kubesphere.io/docs/installing-on-linux/public-cloud/install-kubesphere-on-ali-ecs/)
- [Installing on Huaweicloud VM](https://https://v3-0.docs.kubesphere.io/docs/installing-on-linux/public-cloud/install-ks-on-huaweicloud-ecs/)

## Contributing, Support, Discussion, and Community

We :heart: your contribution. The [community](https://github.com/kubesphere/community) walks you through how to get started contributing KubeSphere. The [development guide](https://github.com/kubesphere/community/tree/master/developer-guide/development) explains how to set up development environment.

- [Slack Channel](https://join.slack.com/t/kubesphere/shared_invite/enQtNTE3MDIxNzUxNzQ0LTZkNTdkYWNiYTVkMTM5ZThhODY1MjAyZmVlYWEwZmQ3ODQ1NmM1MGVkNWEzZTRhNzk0MzM5MmY4NDc3ZWVhMjE)
- [Youtube](https://www.youtube.com/channel/UCyTdUQUYjf7XLjxECx63Hpw)
- [Follow us on Twitter](https://twitter.com/KubeSphere)

Please submit any KubeSphere bugs, issues, and feature requests to [KubeSphere GitHub Issue](https://github.com/kubesphere/kubesphere/issues).

## Who are using KubeSphere

The [user case studies](https://kubesphere.io/case/) page includes the user list of the project. You can [submit a PR](https://github.com/kubesphere/kubesphere/blob/master/docs/powered-by-kubesphere.md) to add your institution name and homepage if you are using KubeSphere.

## Landscapes

<p align="center">
<br/><br/>
<img src="https://landscape.cncf.io/images/left-logo.svg" width="150"/>&nbsp;&nbsp;<img src="https://landscape.cncf.io/images/right-logo.svg" width="200"/>&nbsp;&nbsp;<img src="https://www.cncf.io/wp-content/uploads/2017/11/certified_kubernetes_color.png" height="40" width="30"/>
<br/><br/>
KubeSphere is a member of CNCF and a <a href="https://www.cncf.io/certification/software-conformance/#logos">Kubernetes Conformance Certified platform
</a>, which enriches the <a href="https://landscape.cncf.io/landscape=observability-and-analysis&license=apache-license-2-0">CNCF CLOUD NATIVE Landscape.
</a>
</p>
