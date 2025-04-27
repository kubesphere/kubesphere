<p align="center">
<a href="https://kubesphere.io/"><img src="docs/images/kubesphere-icon.gif" alt="banner" width="200px"></a>
</p>

<p align="center">
<b>The container platform tailored for <i>Kubernetes multi-cloud, datacenter, and edge</i> management</b>
</p>

<p align=center>
<a href="https://goreportcard.com/report/github.com/kubesphere/kubesphere"><img src="https://goreportcard.com/badge/github.com/kubesphere/kubesphere" alt="A+"></a>
<a href="https://hub.docker.com/r/kubesphere/ks-installer"><img src="https://img.shields.io/docker/pulls/kubesphere/ks-installer"></a>
<a href="https://github.com/kubesphere/kubesphere/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22good+first+issue%22"><img src="https://img.shields.io/github/issues/kubesphere/kubesphere/good%20first%20issue?logo=github" alt="good first issue"></a>
<a href="https://twitter.com/intent/follow?screen_name=KubeSphere"><img src="https://img.shields.io/twitter/follow/KubeSphere?style=social" alt="follow on Twitter"></a>
<a href="https://join.slack.com/t/kubesphere/shared_invite/zt-2b4t6rdb4-ico_4UJzCln_S2c1pcrIpQ"><img src="https://img.shields.io/badge/Slack-2000%2B-blueviolet?logo=slack&amp;logoColor=white"></a>
<a href="https://www.youtube.com/channel/UCyTdUQUYjf7XLjxECx63Hpw"><img src="https://img.shields.io/youtube/channel/subscribers/UCyTdUQUYjf7XLjxECx63Hpw?style=social"></a>
</p>


----

## What is KubeSphere

> English | [中文](README_zh.md)

[KubeSphere](https://kubesphere.io/) is a **distributed operating system for cloud-native application management**,
using [Kubernetes](https://kubernetes.io) as its kernel. It provides a plug-and-play architecture, allowing third-party
applications to be seamlessly integrated into its ecosystem. KubeSphere is also a multi-tenant container platform with
full-stack automated IT operation and streamlined DevOps workflows. It provides developer-friendly wizard web UI,
helping enterprises to build out a more robust and feature-rich platform, which includes most common functionalities
needed for enterprise Kubernetes strategy, see [Feature List](#features) for details.

The following screenshots give a close insight into KubeSphere. Please
check [What is KubeSphere](https://kubesphere.io/docs/introduction/what-is-kubesphere/) for further information.

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

## Demo environment

🎮 [KubeSphere Lite](https://kubesphere.cloud/en/console/managed-cluster/) provides you with free, stable, and
out-of-the-box managed cluster service. After registration and login, you can easily create a K8s cluster with
KubeSphere installed in only 5 seconds and experience feature-rich KubeSphere.

🖥 You can view the [Demo Video](https://youtu.be/YxZ1YUv0CYs) to get started with KubeSphere.

## Features

<details>
<summary><b>🧩 Extensible Architecture</b></summary>  
Designed for flexibility, supporting plugin-based extensions and seamless integrations. Easily customize and expand functionalities to meet evolving needs. <a href="https://kubesphere.io/docs/v4.1/01-intro/02-architecture/">Learn more</a>.  
</details>

<details>
  <summary><b>🕸 Provisioning Kubernetes Cluster</b></summary>
  Support deploy Kubernetes on any infrastructure, support online and air-gapped installation. <a href="https://kubesphere.io/docs/v4.1/03-installation-and-upgrade/02-install-kubesphere/">Learn more</a>.
  </details>

<details>
  <summary><b>🔗 Kubernetes Multi-cluster Management</b></summary>
  Provide a centralized control plane to manage multiple Kubernetes clusters, and support the ability to propagate an app to multiple K8s clusters across different cloud providers.
  </details>

<details>
  <summary><b>🤖 Kubernetes DevOps</b></summary>
  Provide GitOps-based CD solutions and use Argo CD to provide the underlying support, collecting CD status information in real time. With the mainstream CI engine Jenkins integrated, DevOps has never been easier. <a href="https://kubesphere.io/docs/v4.1/11-use-extensions/01-devops/01-overview/">Learn more</a>.
  </details>

<details>
  <summary><b>🔎 Cloud Native Observability</b></summary>
  Multi-dimensional monitoring, events and auditing logs are supported; multi-tenant log query and collection, alerting and notification are built-in. <a href="https://kubesphere.io/docs/v4.1/11-use-extensions/05-observability-platform/">Learn more</a>.
  </details>

<details>
  <summary><b>🌐 Service Mesh (Istio-based)</b></summary>
  Provide fine-grained traffic management, observability and tracing for distributed microservice applications, provides visualization for traffic topology. <a href="https://kubesphere.io/docs/v4.1/11-use-extensions/03-service-mesh/">Learn more</a>.
  </details>

<details>
  <summary><b>💻 App Store</b></summary>
  Provide an App Store for Helm-based applications, and offer application lifecycle management on Kubernetes platform. <a href="https://kubesphere.io/docs/v4.1/11-use-extensions/02-app-store/02-app-management/">Learn more</a>.
  </details>

<details>
  <summary><b>💡 Edge Computing Platform</b></summary>
  KubeSphere integrates <a href="https://kubeedge.io/en/">KubeEdge</a> to enable users to deploy applications on the edge devices and view logs and monitoring metrics of them on the console. <a href="https://kubesphere.io/docs/v4.1/11-use-extensions/17-kubeedge/">Learn more</a>.
  </details>

<details>
  <summary><b>🗃 Support Multiple Storage and Networking Solutions</b></summary>
  <li>Support GlusterFS, CephRBD, NFS, LocalPV solutions, and provide CSI plugins to consume storage from multiple cloud providers.</li><li>Provide Load Balancer Implementation <a href="https://github.com/kubesphere/openelb">OpenELB</a> for Kubernetes in bare-metal, edge, and virtualization.</li><li> Provides network policy and Pod IP pools management, support Calico, Flannel, Kube-OVN</li>.</li>.
  </details>

<details>
<summary><b>🏢 Multi-Tenancy</b></summary>  
Isolated workspaces with role-based access control ensure secure resource sharing across multiple tenants. Supports fine-grained permissions and quota management. <a href="https://kubesphere.io/docs/v4.1/08-workspace-management/">Learn more</a>.  
</details>

<details>
  <summary><b>🧠 GPU Workloads Scheduling and Monitoring</b></summary>
  Create GPU workloads on the GUI, schedule GPU resources, and manage GPU resource quotas by tenant.
  </details>

## Architecture

KubeSphere 4.x adopts a microkernel + extension components architecture ([codename LuBan](https://kubesphere.io/docs/v4.1/01-intro/01-introduction/)). The core part (KubeSphere Core) only includes the essential basic functions required for system operation, with independent functional modules split and provided in the form of extension components. Users can dynamically manage the extension components during system operation. With the extension capabilities, KubeSphere can support more application scenarios and meet the needs of different users.

![Architecture](docs/images/architecture.png)

----

## Latest release

🎉 KubeSphere v4.1.2 was released! It brings enhancements and better user experience, see
the [Release Notes For 4.1.2](https://kubesphere.io/docs/v4.1/20-release-notes/release-v412/) for the updates.

## Installation

KubeSphere can run anywhere from on-premise datacenter to any cloud to edge. In addition, it can be deployed on any
version-compatible Kubernetes cluster. KubeSphere consumes very few resources, and you can
optionally [install additional extensions after installation](https://kubesphere.io/docs/v4.1/02-quickstart/03-install-an-extension/).

### Quick start

#### Installing on K8s

Run the following commands to install KubeSphere on an existing Kubernetes cluster:

```bash
helm upgrade --install -n kubesphere-system --create-namespace ks-core https://charts.kubesphere.io/main/ks-core-1.1.3.tgz --debug --wait
```

### KubeSphere for hosted Kubernetes services

KubeSphere is hosted on the following cloud providers, and you can try KubeSphere by one-click installation on their
hosted Kubernetes services.

- [KubeSphere for Amazon EKS](https://aws.amazon.com/quickstart/architecture/qingcloud-kubesphere/)
- [KubeSphere for Azure AKS](https://market.azure.cn/marketplace/apps/qingcloud.kubesphere)
- [KubeSphere for DigitalOcean Kubernetes](https://marketplace.digitalocean.com/apps/kubesphere)
- [KubeSphere on QingCloud AppCenter(QKE)](https://www.qingcloud.com/products/kubesphereqke)

You can also install KubeSphere on other hosted Kubernetes services within minutes, see
the [step-by-step guides](https://kubesphere.io/docs/v4.1/02-quickstart/01-install-kubesphere/) to get started.

> 👨‍💻 No internet access? Refer to
> the [Air-gapped Installation](https://kubesphere.io/docs/v4.1/03-installation-and-upgrade/02-install-kubesphere/04-offline-installation/).

## Guidance, discussion, contribution, and support

You can reach the KubeSphere [community](https://github.com/kubesphere/community) and developers via the following
channels:

- [Slack Channel](https://join.slack.com/t/kubesphere/shared_invite/zt-2b4t6rdb4-ico_4UJzCln_S2c1pcrIpQ)
- [Youtube](https://www.youtube.com/channel/UCyTdUQUYjf7XLjxECx63Hpw)
- [X/Twitter](https://x.com/KubeSphere)

:hugs: Please submit any KubeSphere bugs, issues, and feature requests
to [KubeSphere GitHub Issue](https://github.com/kubesphere/kubesphere/issues).

:heart_decoration: The KubeSphere team also provides efficient official ticket support to respond in hours. For more
information, click [KubeSphere Online Support](https://kubesphere.cloud/en/ticket/).

## Contribution

- [KubeSphere Development Guide](https://github.com/kubesphere/community/tree/master/developer-guide/development)
  explains how to build and develop KubeSphere.
- [KubeSphere Extension Development Guide](https://dev-guide.kubesphere.io/extension-dev-guide/en/) explains how to
  develop KubeSphere extensions.

## Code of conduct

Participation in the KubeSphere community is governed by
the [Code of Conduct](https://github.com/kubesphere/community/blob/master/code-of-conduct.md).

## Security

The security process for reporting vulnerabilities is described in [SECURITY.md](./SECURITY.md).

## Who are using KubeSphere

The [user case studies](https://kubesphere.io/case/) page includes the user list of the project. You
can [leave a comment](https://github.com/kubesphere/kubesphere/issues/4123) to let us know your use case.

---

<p align="center">
<br/><br/>
<img src="https://raw.githubusercontent.com/cncf/artwork/refs/heads/main/other/cncf-landscape/horizontal/color/cncf-landscape-horizontal-color.svg" width="150"/>&nbsp;&nbsp;<img src="https://raw.githubusercontent.com/cncf/artwork/refs/heads/main/other/cncf/horizontal/color/cncf-color.svg" width="200"/>&nbsp;&nbsp;
<br/><br/>
KubeSphere is a member of CNCF and a <a href="https://www.cncf.io/certification/software-conformance/#logos">Kubernetes Conformance Certified platform
</a>, which enriches the <a href="https://landscape.cncf.io/?landscape=observability-and-analysis&group=certified-partners-and-providers&item=platform--certified-kubernetes-distribution--kubesphere">CNCF CLOUD NATIVE Landscape.
</a>
</p>
