<p align="center">
<a href="https://kubesphere.com.cn/"><img src="docs/images/kubesphere-icon.gif" alt="banner" width="200px"></a>
</p>

<p align="center">
<b>适用于<i> Kubernetes 多云、数据中心和边缘 </i>管理的容器平台</b>
</p>

<p align=center>
<a href="https://goreportcard.com/report/github.com/kubesphere/kubesphere"><img src="https://goreportcard.com/badge/github.com/kubesphere/kubesphere" alt="A+"></a>
<a href="https://hub.docker.com/r/kubesphere/ks-installer"><img src="https://img.shields.io/docker/pulls/kubesphere/ks-installer"></a>
<a href="https://github.com/kubesphere/kubesphere/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22good+first+issue%22"><img src="https://img.shields.io/github/issues/kubesphere/kubesphere/good%20first%20issue?logo=github" alt="good first issue"></a>
<a href="https://twitter.com/intent/follow?screen_name=KubeSphere"><img src="https://img.shields.io/twitter/follow/KubeSphere?style=social" alt="follow on Twitter"></a>
<a href="https://join.slack.com/t/kubesphere/shared_invite/zt-219hq0b5y-el~FMRrJxGM1Egf5vX6QiA"><img src="https://img.shields.io/badge/Slack-2000%2B-blueviolet?logo=slack&amp;logoColor=white"></a>
<a href="https://www.youtube.com/channel/UCyTdUQUYjf7XLjxECx63Hpw"><img src="https://img.shields.io/youtube/channel/subscribers/UCyTdUQUYjf7XLjxECx63Hpw?style=social"></a>
</p>


----

## KubeSphere 是什么

> [English](README.md) | 中文

[KubeSphere](https://kubesphere.io/zh/) 愿景是打造一个以 [Kubernetes](https://kubernetes.io/zh/) 为内核的 **云原生分布式操作系统**，它的架构可以非常方便地使第三方应用与云原生生态组件进行即插即用（plug-and-play）的集成，支持云原生应用在多云与多集群的统一分发和运维管理。 KubeSphere 也是一个多租户容器平台，提供全栈的 IT 自动化运维的能力，简化企业的 DevOps 工作流。KubeSphere 提供了运维友好的向导式操作界面，帮助企业快速构建一个强大和功能丰富的容器云平台，详情请参阅 [平台功能](#平台功能) 。

下面的屏幕截图让我们进一步了解 KubeSphere，关于 KubeSphere 更详细的介绍与说明请参阅 [什么是 KubeSphere](https://kubesphere.io/zh/docs/introduction/what-is-kubesphere/) 。

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

## Demo 环境

🎮 [KubeSphere Cloud 轻量集群](https://kubesphere.cloud/console/managed-cluster/)为您提供免费、稳定且开箱即用的 KubeSphere 托管集群服务。注册账号并登录后，可在 5 秒内新建一个安装 KubeSphere 的 K8s 集群，进而进入 KubeSphere 交互式体验各项功能。

🖥 您还可以通过 [Demo 视频](https://youtu.be/YxZ1YUv0CYs)快速了解使用操作。

## 平台功能

<details>
<summary><b>🧩 可扩展架构</b></summary>  
旨在提供灵活性，支持基于插件的扩展和无缝集成。轻松定制和扩展功能，以满足不断变化的需求， <a href="https://kubesphere.io/docs/v4.1/01-intro/02-architecture/">了解更多</a>.  
</details>


<details>
  <summary><b>🕸 部署 Kubernetes 集群</b></summary>
  支持在任何基础设施上部署 Kubernetes，支持在线安装和离线安装，<a href="https://kubesphere.io/docs/v4.1/03-installation-and-upgrade/02-install-kubesphere/">了解更多</a>。
  </details>

<details>
  <summary><b>🔗 Kubernetes 多集群管理</b></summary>
  提供集中控制平台来管理多个 Kubernetes 集群，支持将应用程序发布到跨不同云供应商的多个 k8s 集群上。
  </details>

<details>
  <summary><b>🤖 Kubernetes DevOps</b></summary>
  提供基于 GitOps 的 CD 方案，底层支持 Argo CD，可实时统计 CD 状态。结合主流 CI 引擎 Jenkins，让 DevOps 更加易用。<a href="https://kubesphere.io/docs/v4.1/11-use-extensions/01-devops/01-overview/">了解更多</a>。
  </details>

<details>
  <summary><b>🔎 云原生可观测性</b></summary>
  支持多维度监控、事件和审计日志；内置多租户日志查询和收集，告警和通知，<a href="https://kubesphere.io/docs/v4.1/11-use-extensions/05-observability-platform/">了解更多</a>。
  </details>

<details>
  <summary><b>🌐 基于 Istio 的微服务治理</b></summary>
  为分布式微服务应用程序提供细粒度的流量管理、可观测性和服务跟踪，支持可视化的流量拓扑，<a href="https://kubesphere.io/docs/v4.1/11-use-extensions/03-service-mesh/">了解更多</a>。
  </details>

<details>
  <summary><b>💻 应用商店</b></summary>
  为基于 Helm 的应用程序提供应用商店，并在 Kubernetes 平台上提供应用程序生命周期管理功能，<a href="https://kubesphere.io/docs/v4.1/11-use-extensions/02-app-store/02-app-management/">了解更多</a>。
  </details>

<details>
  <summary><b>💡 Kubernetes 边缘节点管理</b></summary>
  基于 <a href="https://kubeedge.io/zh/">KubeEdge</a> 实现应用与工作负载在云端与边缘节点的统一分发与管理，解决在海量边、端设备上完成应用交付、运维、管控的需求，<a href= "https://kubesphere.io/docs/v4.1/11-use-extensions/17-kubeedge/">了解更多</a>。
  </details>

<details>
  <summary><b>🗃 支持多种存储和网络解决方案</b></summary>
  <li>支持 GlusterFS、CephRBD、NFS、LocalPV ，并提供多个 CSI 插件对接公有云与企业级存储。</li><li>提供 Kubernetes 在裸机、边缘和虚拟化中的负载均衡器实现 <a href="https://github.com/kubesphere/openelb">OpenELB</a> 。</li><li>提供网络策略和容器组 IP 池管理，支持 Calico、Flannel、Kube-OVN。</li>
  </details>

<details>
  <summary><b>🏘 多租户与统一鉴权认证</b></summary>
  具有基于角色的访问控制的逻辑隔离可确保跨多个租户的安全资源共享。支持细粒度的权限和配额管理，<a href="https://kubesphere.io/docs/v4.1/08-workspace-management/">了解更多</a>。
  </details>

<details>
  <summary><b>🧠 GPU 工作负载调度与监控</b></summary>
  支持可视化创建 GPU 工作负载，支持 GPU 监控，同时还支持对 GPU 资源进行租户级配额管理。
  </details>

## 架构说明

KubeSphere 4.x，采用了微内核 + 扩展组件的架构（[代号 LuBan](https://kubesphere.io/docs/v4.1/01-intro/01-introduction/)）。其中内核部分（KubeSphere Core）仅包含系统运行的必备基础功能，将独立的功能模块拆分通过扩展组件（Extensions）的形式提供。用户可在系统运行时动态地管理扩展组件，借助扩展能力，KubeSphere 可以支持更多的应用场景，满足不同用户的需求。

![Architecture](docs/images/architecture.png)

----

## 最新版本

🎉 KubeSphere 4.1.2 全新发布！！多项功能与体验优化，带来更好的产品体验，详见 [v4.1.2 版本说明](https://kubesphere.io/docs/v4.1/20-release-notes/release-v412/) 。

## 安装

KubeSphere 支持在任意平台运行，从本地数据中心到混合多云再走向边缘。此外，KubeSphere 可以部署在任何版本兼容的 Kubernetes 集群上。KubeSphere 的资源消耗很少, 你可以在安装完成后[安装其他的扩展组件](https://kubesphere.io/docs/v4.1/02-quickstart/03-install-an-extension/)。

### 快速入门
#### 在 K8s 上安装

运行以下命令以在现有 Kubernetes 集群上安装 KubeSphere：

```bash
helm upgrade --install -n kubesphere-system --create-namespace ks-core https://charts.kubesphere.io/main/ks-core-1.1.3.tgz --debug --wait
```

### 在托管 Kubernetes 上部署 KubeSphere

KubeSphere 托管在以下云供应商上，您可以通过在其托管的 Kubernetes 服务上一键安装来部署 KubeSphere。

- [在 Amazon EKS 上部署 KubeSphere](https://aws.amazon.com/quickstart/architecture/qingcloud-kubesphere/)
- [在 Azure AKS 上部署 KubeSphere](https://market.azure.cn/marketplace/apps/qingcloud.kubesphere)
- [在 DigitalOcean 上部署 KubeSphere](https://marketplace.digitalocean.com/apps/kubesphere)
- [在青云QingCloud QKE 上部署 KubeSphere](https://www.qingcloud.com/products/kubesphereqke)

您还可以在几分钟内在其他托管的 Kubernetes 服务上安装 KubeSphere，请参阅 [官方文档](https://kubesphere.io/zh/docs/installing-on-kubernetes/) 以开始使用。

> 👨‍💻 不能访问网络？参考[在离线环境中安装 KubeSphere](https://kubesphere.io/docs/v4.1/03-installation-and-upgrade/02-install-kubesphere/04-offline-installation/)。

## 指引、讨论、贡献与支持

我们 :heart: 您的贡献。[社区](https://github.com/kubesphere/community) 将引导您了解如何开始贡献 KubeSphere。[开发指南](https://github.com/kubesphere/community/tree/master/developer-guide/development) 说明了如何安装开发环境。

- [中文论坛](https://kubesphere.com.cn/forum/)
- [社区微信群（见官网底部）](https://kubesphere.com.cn/)
- [Slack Channel](https://join.slack.com/t/kubesphere/shared_invite/zt-219hq0b5y-el~FMRrJxGM1Egf5vX6QiA)
- [Bilibili](https://space.bilibili.com/438908638)
- [X/Twitter](https://twitter.com/KubeSphere)

:hugs: 请将任何 KubeSphere 的 Bug、问题和需求提交到 [KubeSphere GitHub Issue](https://github.com/kubesphere/kubesphere/issues)。

:heart_decoration: 若您期待官方、高效的技术服务，青云科技也为 KubeSphere 开源版本提供全程可靠、小时响应的工单支持，详情垂询 [KubeSphere 在线技术支持](https://kubesphere.cloud/ticket/)。

## 谁在使用 KubeSphere

[用户案例学习](https://kubesphere.io/zh/case/) 列出了哪些企业在使用 KubeSphere。欢迎 [发表评论](https://github.com/kubesphere/kubesphere/issues/4123) 来分享您的使用案例。

## Landscapes

<p align="center">
<br/><br/>
<img src="https://raw.githubusercontent.com/cncf/artwork/refs/heads/main/other/cncf-landscape/horizontal/color/cncf-landscape-horizontal-color.svg" width="150"/>&nbsp;&nbsp;<img src="https://raw.githubusercontent.com/cncf/artwork/refs/heads/main/other/cncf/horizontal/color/cncf-color.svg" width="200"/>&nbsp;&nbsp;
<br/><br/>
KubeSphere 是 CNCF 基金会成员并且通过了 <a href="https://www.cncf.io/certification/software-conformance/#logos">Kubernetes 一致性认证
</a>，进一步丰富了 <a href="https://landscape.cncf.io/?landscape=observability-and-analysis&license=apache-license-2-0">CNCF 云原生的生态。
</a>
</p>
