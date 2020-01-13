# KubeSphere
[![License](http://img.shields.io/badge/license-apache%20v2-blue.svg)](https://github.com/KubeSphere/KubeSphere/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/kubesphere/kubesphere.svg?branch=master)](https://travis-ci.org/kubesphere/kubesphere)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubesphere/kubesphere)](https://goreportcard.com/report/github.com/kubesphere/kubesphere)
[![KubeSphere release](https://img.shields.io/github/release/kubesphere/kubesphere.svg?color=release&label=release&logo=release&logoColor=release)](https://github.com/kubesphere/kubesphere/releases/tag/v2.1.0)

![logo](docs/images/kubesphere-logo.png)

----

## What is KubeSphere

> English | [中文](README_zh.md)

[KubeSphere](https://kubesphere.io/) is a multi-tenant enterprise-grade container platform built on [Kubernetes](http://kubernetes.io), with full-stack automated IT operation and streamlined DevOps workflows. KubeSphere provides developer-friendly wizard web UI, helps enterprises to build out a more robust and feature-rich platform, includes most common functionalities needed for enterprise Kubernetes strategy, such as the **Kubernetes resource management, DevOps (CI/CD), application lifecycle management, monitoring, logging, Service Mesh (Istio-based), multi-tenancy, alerting and notification, storage and networking, autoscaling, access control, GPU support, etc.**, as well as **multi-cluster management, Network Policy, registry management, security** in upcoming releases. KubeSphere is going to be a **distributed operating system with cloud native stack** based on Kubernetes, will be very well architected for **plug-and-play integration with its ecosystem** as well.

KubeSphere provides a complete user experience around Kubernetes that incorporates a rich set of cloud native ecosystem tools, allows developers and DevOps teams use their favorite tools in a single front-end interface. KubeSphere delivers **consolidated views while integrating a wide breadth of ecosystem tools** upon Kubernetes and offer **consistent user experience** to reduce complexity. Most importantly, these functionalities are loosely coupled with the platform since they are pluggable and optional based on your demands, will not impact the flexibilty of Kubernetes.

![](docs/images/kubesphere-platform-overview.png)


> Note: The [Screenshots](docs/en/guides/screenshots.md) give a close insight into KubeSphere, see [What is KubeSphere](https://kubesphere.io/docs/v2.1/en/introduction/what-is-kubesphere) for details.


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

## Demo Environment

Using the account `demo1 / Demo123` to log in the [demo environment](https://demo.kubesphere.io/). Please note the account is granted viewer access.

## Features

KubeSphere provides an easy-to-use console with awesome user experience that allows you to quickly get started with a container management platform. KubeSphere provides and supports the following major features:


- Workload management
- Service mesh (Istio-based)
- DevOps (CI/CD Pipeline)
- Source to Image, Binary to Image
- Multi-tenant management
- Multi-dimensional and multi-tenant monitoring, logging, alerting, notification
- Service and network management
- Application store and application lifecycle management
- Node and storage class management, image registry management
- Integrated Harbor, GitLab, SonarQube
- LB controller for Kubernetes on bare metal ([Porter](https://github.com/kubesphere/porter)), [cloud LB plugin](https://github.com/yunify/qingcloud-cloud-controller-manager)
- Support GPU node, support [vGPU](https://github.com/virtaitech/orion)


It also supports a variety of open source storage solutions and cloud storage products as the persistent storage services, as well as supports multiple open source network plugins.

> Note: See this [document](https://kubesphere.io/docs/v2.1/en/introduction/features/) which elaborates on the KubeSphere features and services.

----

## Architecture

KubeSphere uses a loosely-coupled architecture that separates the [frontend](https://github.com/kubesphere/console) from the [backend](https://github.com/kubesphere/kubesphere), the back end can also be connected with external systems through the REST API, all components are designed as Docker containers. See [Architecture](https://kubesphere.io/docs/v2.1/en/introduction/architecture/) for details.


## Latest Release

KubeSphere 2.1.0 was released on **November 12nd, 2019**. Check the [Release Notes For 2.1.0](https://kubesphere.io/docs/v2.1/zh-CN/release/release-v210/) for the updates.

## Installation

KubeSphere can run anywhere from on-premise datacenter to any cloud to edge. In addition, it can be deployed on any Kubernetes distribution.

> Attention: The following section is only used for minimal installation by default, see [Complete Installation Guide](https://kubesphere.io/docs/v2.1/en/installation/intro/) for details.

### Deploy on Existing Kubernetes

**Prerequisites**

> - `Kubernetes version`: `1.13.0 ≤ K8s version < 1.16`;
> - `Helm version`: `2.10.0 ≤ Helm ＜ 3.0.0`，(will support Helm v3 in KubeSphere v3.0) see [Install and Configure Helm in Kubernetes](https://devopscube.com/install-configure-helm-kubernetes/);
> - CPU > 1 Core，Memory > 2 G;
> - An existing Storage Class in your Kubernetes clusters, use `kubectl get sc` to verify it.

Run the following command. When all Pods of KubeSphere are running, it means the installation is successsful. Then you can use `http://<IP>:30880` to access the dashboard with default account `admin/P@88w0rd`.

```yaml
$ kubectl apply -f https://raw.githubusercontent.com/kubesphere/ks-installer/master/kubesphere-minimal.yaml
```


### Deploy on Linux

KubeSphere Installer can help you to install KubeSphere and Kubernetes on your linux machines. It provides [All-in-One](https://kubesphere.com.cn/docs/v2.1/en/installation/all-in-one/) and [Multi-Node](https://kubesphere.com.cn/docs/v2.1/en/installation/multi-node/) installation options.

**Prerequisites**

- Operating Systems
   - CentOS 7.5 (64 bit)
   - Ubuntu 16.04/18.04 LTS (64 bit)
   - Red Hat Enterprise Linux Server 7.4 (64 bit)
   - Debian Stretch 9.5 (64 bit)
- Hardware
   - CPU：2 Core,  Memory：4 G, Disk Space：100 G

##### All-in-One

For those who are new to KubeSphere and looking for the easiest way to install and experience the dashboard. Execute the following commands to download and install KubeSphere in a single node.

```bash
$ curl -L https://kubesphere.io/download/stable/v2.1.0 > installer.tar.gz \
&& tar -zxf installer.tar.gz && cd kubesphere-all-v2.1.0/scripts
$ ./install.sh
```

Choose `"1) All-in-one"` to start the installation without changing any configuration.

> Note: In a development or production environment, it's highly recommended to install Multi-Node KubeSphere.


## To start using KubeSphere

### Quick Start

KubeSphere provides 12 quick-start tutorials to walk you through the platform.

- [Get Started - En](https://kubesphere.io/docs/v2.1/en/quick-start/admin-quick-start/)
- [Get Started - 中](https://kubesphere.io/docs/v2.1/zh-CN/quick-start/admin-quick-start/)


### Documentation

- KubeSphere Documentation ([En](https://kubesphere.io/docs/en/)/[中](https://kubesphere.io/docs/zh-CN/)）
- [API Documentation](https://kubesphere.io/docs/v2.1/en/api-reference/api-docs/)


## To start developing KubeSphere

The [development guide](CONTRIBUTING.md) hosts all information about building KubeSphere from source, git workflow, how to contribute code and how to test.

## RoadMap

Currently, KubeSphere has released the following 4 major editions. The future releases will include Multicluster, Big data, AI, SDN, etc. See [Plans for 2.1.1 and 3.0.0](https://github.com/kubesphere/kubesphere/issues/1368) for more details.

**Express Edition** => **v1.0.x** => **v2.0.x**  => **v2.1.0** => **v2.1.1** => **v3.0.0**

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

## Who Uses KubeSphere

The [Powered by KubeSphere](docs/powered-by-kubesphere.md) page includes users list of the project. You can submit your institution name and homepage if you are using KubeSphere.


## Support, Discussion, and Community

If you need any help with KubeSphere, please join us at [Slack Channel](https://join.slack.com/t/kubesphere/shared_invite/enQtNTE3MDIxNzUxNzQ0LTZkNTdkYWNiYTVkMTM5ZThhODY1MjAyZmVlYWEwZmQ3ODQ1NmM1MGVkNWEzZTRhNzk0MzM5MmY4NDc3ZWVhMjE).

Please submit any KubeSphere bugs, issues, and feature requests to [KubeSphere GitHub Issue](https://github.com/kubesphere/kubesphere/issues).

## Contributing to the project

This [document](docs/en/guides/README.md) walks you through how to get started contributing KubeSphere.
