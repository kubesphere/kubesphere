# KubeSphere 容器平台

[![License](http://img.shields.io/badge/license-apache%20v2-blue.svg)](https://github.com/KubeSphere/KubeSphere/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/kubesphere/kubesphere.svg?branch=master)](https://travis-ci.org/kubesphere/kubesphere)
[![KubeSphere release](https://img.shields.io/github/release/kubesphere/kubesphere.svg?color=release&label=release&logo=release&logoColor=release)](https://github.com/kubesphere/kubesphere/releases/tag/v3.0.0)

![logo](docs/images/kubesphere-logo.png)

----

## KubeSphere 是什么

> [English](README.md) | 中文

[KubeSphere](https://kubesphere.com.cn) 是在 [Kubernetes](https://kubernetes.io) 之上构建的面向云原生应用的 **容器混合云**，支持多云与多集群管理，提供全栈的 IT 自动化运维的能力，简化企业的 DevOps 工作流。KubeSphere 提供了运维友好的向导式操作界面，帮助企业快速构建一个强大和功能丰富的容器云平台。KubeSphere 愿景是打造一个基于 Kubernetes 的云原生分布式操作系统，它的架构可以很方便地与云原生生态进行即插即用（plug-and-play）的集成。

KubeSphere 目前最新的版本为 3.0.0，所有版本 100% 开源，关于 KubeSphere 更详细的介绍与说明请参阅 [什么是 KubeSphere](https://kubesphere.com.cn/docs/zh-CN/introduction/what-is-kubesphere/)。

<table>
  <tr>
      <td width="50%" align="center"><b>工作台</b></td>
      <td width="50%" align="center"><b>项目资源</b></td>
  </tr>
  <tr>
     <td><img src="docs/images/console.png"/></td>
     <td><img src="docs/images/project.png"/></td>
  </tr>
  <tr>
      <td width="50%" align="center"><b>CI/CD 流水线</b></td>
      <td width="50%" align="center"><b>应用商店</b></td>
  </tr>
  <tr>
     <td><img src="docs/images/cicd.png"/></td>
     <td><img src="docs/images/app-store.png"/></td>
  </tr>
</table>

## 快速体验

使用体验账号 `demo1 / Demo123` 登录 [Demo 环境](https://demo.kubesphere.io/)，该账号仅授予了 view 权限，建议自行安装体验完整的管理功能。您还可以访问 Youtube 查看 [KubeSphere Demo 视频](https://youtu.be/u5lQvhi_Xlc)。

## 架构

KubeSphere 采用了前后端分离的架构设计，后端的各个功能组件可通过 REST API 对接外部系统，详见 [架构说明](https://kubesphere.com.cn/docs/zh-CN/introduction/architecture/)。本仓库仅包含后端代码，前端代码参考 [Console 项目](https://github.com/kubesphere/console)。

![Architecture](docs/images/architecture.png)

## 核心功能

|功能 |介绍 |
| --- | ---|
|多云与多集群管理|提供多云与多集群的中央管理面板，支持集群导入，支持应用在多云与多集群一键分发|
| Kubernetes 集群搭建与运维 | 支持在线 & 离线安装、升级与扩容 K8s 集群，支持安装 “云原生全家桶” |
| Kubernetes 资源可视化管理 | 可视化纳管原生 Kubernetes 资源，支持向导式创建与管理 K8s 资源 |
| 基于 Jenkins 的 DevOps 系统 | 支持图形化与脚本两种方式构建 CI/CD 流水线，内置 Source to Image（S2I）和 Binary to Image（B2I）等 CD 工具 |
| 应用商店与应用生命周期管理 | 提供应用商店，内置 Redis、MySQL 等 15 个常用应用，支持应用的生命周期管理 |
| 基于 Istio 的微服务治理 (Service Mesh) | 提供可视化无代码侵入的 **灰度发布、熔断、流量治理与流量拓扑、分布式 Tracing** |
| 多租户管理 | 提供基于角色的细粒度多租户统一认证，支持 **对接企业 LDAP/AD**，提供多层级的权限管理 |
| 丰富的可观察性功能 | 提供集群/工作负载/Pod/容器等多维度的监控，提供基于多租户的日志查询与日志收集，支持节点与应用层级的告警与通知 |
|基础设施管理|支持 Kubernetes 节点管理，支持节点扩容与集群升级，提供基于节点的多项监控指标与告警规则 |
| 存储管理 | 支持对接 Ceph、GlusterFS、NFS、Local PV，支持可视化运维管理 PVC、StorageClass，提供 CSI 插件对接云平台存储 |
| 网络管理 | 提供租户网络隔离与 K8s [Network Policy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) 管理，支持 Calico、Flannel，提供 [Porter LB](https://github.com/kubesphere/porter) 用于暴露物理环境 K8s 集群的 LoadBalancer 服务 |
| GPU support | 集群支持添加 GPU 与 vGPU，可运行 TensorFlow 等 ML 框架 |

以上功能说明详见 [产品功能](https://kubesphere.com.cn/docs/zh-CN/introduction/features/)。

----

## 最新发布

KubeSphere 3.0.0 已于 2020 年 8 月 31 日正式 GA！点击 [Release Notes For 3.0.0](https://kubesphere.com.cn/docs/release/release-v300/) 查看 3.0.0 版本的更新详情。

## 安装 3.0.0

### 快速入门

[快速入门系列](https://kubesphere.com.cn/docs/quick-start/) 提供了快速安装与入门示例，供初次安装体验参考。

### 在已有 Kubernetes 之上安装 KubeSphere

- [基于 Kubernetes 的安装介绍](https://kubesphere.com.cn/docs/installing-on-kubernetes/introduction/overview/)
- [在阿里云 ACK 安装 KubeSphere](https://kubesphere.com.cn/forum/d/1745-kubesphere-v3-0-0-dev-on-ack)
- [在腾讯云 TKE 安装 KubeSphere](https://kubesphere.com.cn/docs/installing-on-kubernetes/hosted-kubernetes/install-ks-on-tencent-tke/)
- [在华为云 CCE 安装 KubeSphere](https://kubesphere.com.cn/docs/installing-on-kubernetes/hosted-kubernetes/install-ks-on-huawei-cce/)
- [在 AWS EKS 安装 KubeSphere](https://kubesphere.com.cn/en/docs/installing-on-kubernetes/hosted-kubernetes/install-kubesphere-on-eks/)
- [在 Google GKE 安装 KubeSphere](https://kubesphere.com.cn/en/docs/installing-on-kubernetes/hosted-kubernetes/install-kubesphere-on-aks/)
- [在 Azure AKS 安装 KubeSphere](https://kubesphere.com.cn/en/docs/installing-on-kubernetes/hosted-kubernetes/install-kubesphere-on-aks/)
- [在 DigitalOcean 安装 KubeSphere](https://kubesphere.com.cn/en/docs/installing-on-kubernetes/hosted-kubernetes/install-kubesphere-on-do/)
- [在 Oracle OKE 安装 KubeSphere](https://kubesphere.com.cn/en/docs/installing-on-kubernetes/hosted-kubernetes/install-kubesphere-on-oke/)

### 基于 Linux 安装 KubeSphere

- [多节点安装介绍（以三节点为例）](https://kubesphere.com.cn/en/docs/installing-on-linux/introduction/multioverview/)
- [在 VMware vSphere 安装高可用集群](https://kubesphere.com.cn/en/docs/installing-on-linux/on-premises/install-kubesphere-on-vmware-vsphere/)
- [在青云QingCloud 安装高可用集群](https://kubesphere.com.cn/en/docs/installing-on-linux/public-cloud/kubesphere-on-qingcloud-instance/)
- [在阿里云 ECS 部署高可用集群](https://kubesphere.com.cn/docs/installing-on-linux/public-cloud/install-kubesphere-on-ali-ecs/)
- [在华为云 VM 部署高可用集群](https://kubesphere.com.cn/docs/installing-on-linux/public-cloud/install-ks-on-huaweicloud-ecs/)
- [在 Azure VM 安装高可用集群](https://kubesphere.com.cn/en/docs/installing-on-linux/public-cloud/install-ks-on-azure-vms/)


## 技术社区

[KubeSphere 社区](https://github.com/kubesphere/community) 包含所有社区的信息，包括如何开发，兴趣小组(SIG)等。比如[开发指南](https://github.com/kubesphere/community/tree/master/developer-guide/development) 详细说明了如何从源码编译、KubeSphere 的 GitHub 工作流、如何贡献代码以及如何测试等。

- [中文论坛](https://kubesphere.com.cn/forum/)
- [Slack Channel](https://join.slack.com/t/kubesphere/shared_invite/enQtNTE3MDIxNzUxNzQ0LTZkNTdkYWNiYTVkMTM5ZThhODY1MjAyZmVlYWEwZmQ3ODQ1NmM1MGVkNWEzZTRhNzk0MzM5MmY4NDc3ZWVhMjE)
- [社区微信群（见官网底部）](https://kubesphere.com.cn/)
- [Bug 与建议反馈（GitHub Issue）](https://github.com/kubesphere/kubesphere/issues)

## 谁在使用 KubeSphere

[Powered by KubeSphere](https://kubesphere.com.cn/case/) 列出了哪些企业在使用 KubeSphere，如果您所在的企业已安装使用了 KubeSphere，欢迎[提交 PR](https://github.com/kubesphere/kubesphere/blob/master/docs/powered-by-kubesphere.md)。

## Landscapes

<p align="center">
<br/><br/>
<img src="https://landscape.cncf.io/images/left-logo.svg" width="150"/>&nbsp;&nbsp;<img src="https://landscape.cncf.io/images/right-logo.svg" width="200"/>&nbsp;&nbsp;<img src="https://www.cncf.io/wp-content/uploads/2017/11/certified_kubernetes_color.png" height="40" width="30"/>
<br/><br/>
KubeSphere 是 CNCF 基金会成员并且通过了 <a href="https://www.cncf.io/certification/software-conformance/#logos">Kubernetes 一致性认证
</a>，进一步丰富了 <a href="https://landscape.cncf.io/landscape=observability-and-analysis&license=apache-license-2-0">CNCF 云原生的生态。
</a>
</p>
