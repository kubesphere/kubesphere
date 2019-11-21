# KubeSphere
[![License](http://img.shields.io/badge/license-apache%20v2-blue.svg)](https://github.com/KubeSphere/KubeSphere/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/kubesphere/kubesphere.svg?branch=master)](https://travis-ci.org/kubesphere/kubesphere)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubesphere/kubesphere)](https://goreportcard.com/report/github.com/kubesphere/kubesphere)
[![KubeSphere release](https://img.shields.io/github/release/kubesphere/kubesphere.svg?color=release&label=release&logo=release&logoColor=release)](https://github.com/kubesphere/kubesphere/releases/tag/advanced-2.0.2)

![logo](docs/images/kubesphere-logo.png)

----

## What is KubeSphere

> English | [中文](README_zh.md)

[KubeSphere](https://kubesphere.io/) is an enterprise-grade multi-tenant container management platform that built on [Kubernetes](https://kubernetes.io). It provides an easy-to-use UI for users to manage computing resources with a few clicks, which reduces the learning curve and empowers the DevOps teams. It greatly reduces the complexity of the daily work of development, testing, operation and maintenance, aiming to alleviate the pain points of Kubernetes' storage, network, security and ease of use, etc.


## Screenshots

> Note: See the [Screenshots](docs/screenshots.md) of KubeSphere to have a most intuitive understanding of KubeSphere dashboard and features.


<table>
  <tr>
      <td width="50%" align="center"><b>KubeSphere Dashboard</b></td>
      <td width="50%" align="center"><b>Project Resources</b></td>
  </tr>
  <tr>
     <td><img src="https://pek3b.qingstor.com/kubesphere-docs/png/20191112094014.png"/></td>
     <td><img src="https://pek3b.qingstor.com/kubesphere-docs/png/20191112094426.png"/></td>
  </tr>
  <tr>
      <td width="50%" align="center"><b>CI/CD Pipeline</b></td>
      <td width="50%" align="center"><b>Application Store</b></td>
  </tr>
  <tr>
     <td><img src="https://pek3b.qingstor.com/kubesphere-docs/png/20190925000712.png"/></td>
     <td><img src="https://pek3b.qingstor.com/kubesphere-docs/png/20191112095006.png"/></td>
  </tr>
</table>

## Video on Youtube

[![KubeSphere](https://pek3b.qingstor.com/kubesphere-docs/png/20191112093503.png)](https://youtu.be/u5lQvhi_Xlc)

## Features

KubeSphere provides an easy-to-use console with the awesome user experience that allows you to quickly get started with a container management platform. KubeSphere provides and supports following core features:


- Workload management
- Service mesh (Istio-based)
- DevOps
- Source to Image
- Multi-tenant management
- Multi-dimensional and Multi-tenant Monitoring, Logging, Alerting, Notification
- Service and network management
- Application template and repository
- Infrastructure management, image registry management
- Integrate Harbor and GitLab
- LB controller for Kubernetes on bare metal ([Porter](https://github.com/kubesphere/porter)), [cloud LB plugin](https://github.com/yunify/qingcloud-cloud-controller-manager)
- Support GPU node


It also supports multiple open source storage and high-performance cloud storage as the persistent storage services, as well as supports multiple open source network plugins.

> Note: See this [document](https://docs.kubesphere.io/advanced-v2.0/zh-CN/introduction/features/) that elaborates on the KubeSphere features and services from a professional point of view.

----

## Architecture

KubeSphere adopts the separation of front and back ends, each component is drawn in the architecture diagram below. KubeSphere can run anywhere from on-premise datacenter to any cloud to edge. In addition, it can be deployed on any Kubernetes distribution.

![](https://pek3b.qingstor.com/kubesphere-docs/png/20190810073322.png)

## Latest Release

KubeSphere 2.1.0 was released on **November 12nd, 2019**. See the [Release Notes For 2.1.0](https://kubesphere.io/docs/v2.1/zh-CN/release/release-v210/) to preview the updates.

## Installation

> Attention: Following section is only used for minimal installation by default, KubeSphere has decoupled some core components in v2.1.0, for more pluggable components installation, see `Enable Pluggable Components` below.

### Deploy On Kubernetes

**Prerequisites**

> - `Kubernetes version`： `1.13.0 ≤ K8s version < 1.16`;
> - `Helm version` >= `2.10.0`，see [Install and Configure Helm in Kubernetes](https://devopscube.com/install-configure-helm-kubernetes/);
> - CPU > 1 Core，Memory > 2 G;
> - An existing Storage Class in your Kubernetes clusters, use `kubectl get sc` to verify it.

When all Pods of KubeSphere are running, it means the installation is successsful. Then you can use `http://IP:30880` to access the dashboard with default account `admin/P@88w0rd`.

```yaml
$ kubectl apply -f https://raw.githubusercontent.com/kubesphere/ks-installer/master/kubesphere-minimal.yaml
```


### Deploy on Linux

- Operating Systems
   - CentOS 7.5 (64 bit)
   - Ubuntu 16.04/18.04 LTS (64 bit)
   - Red Hat Enterprise Linux Server 7.4 (64 bit)
   - Debian Stretch 9.5 (64 bit)
- Hardware
   - CPU：2 Core,  Memory：4 G, Disk Space：100 G

### All-in-One

For those who are new to KubeSphere and looking for the fastest way to install and experience the dashboard. Execute following commands to download and install KubeSphere in a single node.

```bash
$ curl -L https://kubesphere.io/download/stable/v2.1.0 > installer.tar.gz \
&& tar -zxf installer.tar.gz && cd kubesphere-all-v2.1.0/scripts
$ ./install.sh
```

Choose `"1) All-in-one"` to trigger the installation. Generally, you can install it directly without any configuration..

> Note: In a formal environment, it's highly recommended to install KubeSphere with Multi-Node Installation.

### Enable Pluggable Components

The above two methods is only used for minimal installation by default, execute following command to enable more pluggable components installation, make sure your cluster has enough CPU and memory in advance.

```
$ kubectl edit cm -n kubesphere-system ks-installer
```

## To start using KubeSphere

### Quick Start

KubeSphere provides 12 quick-start tutorials to walk you through the process and common manipulation, with a quick overview of the core features of KubeSphere that helps you to get familiar with it.

- [Get Started - En](https://github.com/kubesphere/kubesphere.github.io/tree/master/blog/advanced-2.0/en)
- [Get Started - 中](https://kubesphere.io/docs/advanced-v2.0/zh-CN/quick-start/quick-start-guide/)


### Documentation

- [KubeSphere Documentation (En/中) ](https://kubesphere.io/docs)
- [API Documentation](https://kubesphere.io/docs/advanced-v2.0/zh-CN/api-reference/api-docs/)


## To start developing KubeSphere

The [development guide](CONTRIBUTING.md) hosts all information about building KubeSphere from source, git workflow, how to contribute code and how to test.

## RoadMap

Currently, KubeSphere has released the following 4 major editions. The future releases will include Multicluster, Big data, AI, SDN, etc.

**Express Edition** => **v1.0.x** => **v2.0.x**  => **v2.1.0**

![](https://pek3b.qingstor.com/kubesphere-docs/png/20190926000413.png)

## Landscapes

<p align="center">
<br/><br/>
<img src="https://landscape.cncf.io/images/left-logo.svg" width="150"/>&nbsp;&nbsp;<img src="https://landscape.cncf.io/images/right-logo.svg" width="200"/>&nbsp;&nbsp;<img src="https://www.cncf.io/wp-content/uploads/2017/11/certified_kubernetes_color.png" height="40" width="30"/>
<br/><br/>
KubeSphere is a member of CNCF and a <a href="https://www.cncf.io/certification/software-conformance/#logos">Kubernetes Conformance Certified platform
</a>, which enriches the <a href="https://landscape.cncf.io/landscape=observability-and-analysis&license=apache-license-2-0">CNCF CLOUD NATIVE Landscape.
</a>
</p>


## Support, Discussion, and Community

If you need any help with KubeSphere, please join us at [Slack Channel](https://join.slack.com/t/kubesphere/shared_invite/enQtNTE3MDIxNzUxNzQ0LTZkNTdkYWNiYTVkMTM5ZThhODY1MjAyZmVlYWEwZmQ3ODQ1NmM1MGVkNWEzZTRhNzk0MzM5MmY4NDc3ZWVhMjE).

Please submit any KubeSphere bugs, issues, and feature requests to [KubeSphere GitHub Issue](https://github.com/kubesphere/kubesphere/issues).

## Contributing to the project

All members of the KubeSphere community must abide by [Code of Conduct](docs/code-of-conduct.md). Only by respecting each other can we develop a productive, collaborative community.

How to submit a pull request to KubeSphere? See [Pull Request Instruction](docs/pull-requests.md).
