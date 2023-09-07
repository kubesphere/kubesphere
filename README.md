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
<a href="https://join.slack.com/t/kubesphere/shared_invite/zt-219hq0b5y-el~FMRrJxGM1Egf5vX6QiA"><img src="https://img.shields.io/badge/Slack-2000%2B-blueviolet?logo=slack&amp;logoColor=white"></a>
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
  <summary><b>🕸 Provisioning Kubernetes Cluster</b></summary>
  Support deploy Kubernetes on any infrastructure, support online and air-gapped installation. <a href="https://kubesphere.io/docs/installing-on-linux/introduction/intro/">Learn more</a>.
  </details>

<details>
  <summary><b>🔗 Kubernetes Multi-cluster Management</b></summary>
  Provide a centralized control plane to manage multiple Kubernetes clusters, and support the ability to propagate an app to multiple K8s clusters across different cloud providers.
  </details>

<details>
  <summary><b>🤖 Kubernetes DevOps</b></summary>
  Provide GitOps-based CD solutions and use Argo CD to provide the underlying support, collecting CD status information in real time. With the mainstream CI engine Jenkins integrated, DevOps has never been easier. <a href="https://kubesphere.io/devops/">Learn more</a>.
  </details>

<details>
  <summary><b>🔎 Cloud Native Observability</b></summary>
  Multi-dimensional monitoring, events and auditing logs are supported; multi-tenant log query and collection, alerting and notification are built-in. <a href="https://kubesphere.io/observability/">Learn more</a>.
  </details>

<details>
  <summary><b>🧩 Service Mesh (Istio-based)</b></summary>
  Provide fine-grained traffic management, observability and tracing for distributed microservice applications, provides visualization for traffic topology. <a href="https://kubesphere.io/service-mesh/">Learn more</a>.
  </details>

<details>
  <summary><b>💻 App Store</b></summary>
  Provide an App Store for Helm-based applications, and offer application lifecycle management on Kubernetes platform. <a href="https://kubesphere.io/docs/pluggable-components/app-store/">Learn more</a>.
  </details>

<details>
  <summary><b>💡 Edge Computing Platform</b></summary>
  KubeSphere integrates <a href="https://kubeedge.io/en/">KubeEdge</a> to enable users to deploy applications on the edge devices and view logs and monitoring metrics of them on the console. <a href="https://kubesphere.io/docs/pluggable-components/kubeedge/">Learn more</a>.
  </details>

<details>
  <summary><b>📊 Metering and Billing</b></summary>
  Track resource consumption at different levels on a unified dashboard, which helps you make better-informed decisions on planning and reduce the cost. <a href="https://kubesphere.io/docs/toolbox/metering-and-billing/view-resource-consumption/">Learn more</a>.
  </details>

<details>
  <summary><b>🗃 Support Multiple Storage and Networking Solutions</b></summary>
  <li>Support GlusterFS, CephRBD, NFS, LocalPV solutions, and provide CSI plugins to consume storage from multiple cloud providers.</li><li>Provide Load Balancer Implementation <a href="https://github.com/kubesphere/openelb">OpenELB</a> for Kubernetes in bare-metal, edge, and virtualization.</li><li> Provides network policy and Pod IP pools management, support Calico, Flannel, Kube-OVN</li>.</li>.
  </details>

<details>
  <summary><b>🏘 Multi-tenancy</b></summary>
  Provide unified authentication with fine-grained roles and three-tier authorization system, and support AD/LDAP authentication.
  </details>

<details>
  <summary><b>🧠 GPU Workloads Scheduling and Monitoring</b></summary>
  Create GPU workloads on the GUI, schedule GPU resources, and manage GPU resource quotas by tenant.
  </details>

## Architecture

KubeSphere uses a loosely-coupled architecture that separates the [frontend](https://github.com/kubesphere/console) from
the [backend](https://github.com/kubesphere/kubesphere). External systems can access the components of the backend
through the REST APIs.

![Architecture](docs/images/architecture.png)

----

## Latest release

🎉 KubeSphere v3.4.0 was released! It brings enhancements and better user experience, see
the [Release Notes For 3.4.0](https://kubesphere.io/docs/release/release-v340/) for the updates.

#### Component supported versions table

| Component      | Version                                                                       | K8s supported version         |
|----------------|-------------------------------------------------------------------------------|-------------------------------|
| Alerting       | N/A                                                                           | 1.21,1.22,1.23,1.24,1.25,1.26 |
| Auditing	      | v0.2.0                                                                        | 1.21,1.22,1.23,1.24,1.25,1.26 |
| Monitoring     | N/A		                                                                         | 1.21,1.22,1.23,1.24,1.25,1.26 |
| DevOps         | v3.4.0                                                                        | 1.21,1.22,1.23,1.24,1.25,1.26 |
| EdgeRuntime    | v1.13.0                                                                       | 1.21,1.22,1.23                |
| Events         | N/A                                                                           | 1.21,1.22,1.23,1.24,1.25,1.26 |
| Logging        | opensearch：v2.6.0<br/>fluentbit-operator: v0.14.0<br/> fluent-bit-tag: v1.9.4 | 1.21,1.22,1.23,1.24,1.25,1.26 |
| Metrics Server | v0.4.2                                                                        | 1.21,1.22,1.23,1.24,1.25,1.26 |
| Network        | N/A                                                                           | 1.21,1.22,1.23,1.24,1.25,1.26 |
| Notification   | v2.3.0                                                                        | 1.21,1.22,1.23,1.24,1.25,1.26 |
| AppStore       | N/A                                                                           | 1.21,1.22,1.23,1.24,1.25,1.26 |
| Storage        | pvc-autoresizer: v0.3.0<br/>storageclass-accessor: v0.2.2                     | 1.21,1.22,1.23,1.24,1.25,1.26 |
| ServiceMesh    | Istio: v1.14.6                                                                | 1.21,1.22,1.23,1.24           |
| Gateway        | Ingress NGINX Controller: v1.3.1                                              | 1.21,1.22,1.23,1.24           |


## Installation

KubeSphere can run anywhere from on-premise datacenter to any cloud to edge. In addition, it can be deployed on any
version-compatible Kubernetes cluster. The installer will start a minimal installation by default, you
can [enable other pluggable components before or after installation](https://kubesphere.io/docs/quick-start/enable-pluggable-components/).

### Quick start

#### Installing on K8s/K3s

Ensure that your cluster has installed Kubernetes v1.21.x, v1.22.x, v1.23.x, * v1.24.x, * v1.25.x, or * v1.26.x. For Kubernetes versions with an asterisk, some features may be unavailable due to incompatibility.

Run the following commands to install KubeSphere on an existing Kubernetes cluster:

```yaml
kubectl apply -f https://github.com/kubesphere/ks-installer/releases/download/v3.4.0/kubesphere-installer.yaml

kubectl apply -f https://github.com/kubesphere/ks-installer/releases/download/v3.4.0/cluster-configuration.yaml
```

#### All-in-one

👨‍💻 No Kubernetes? You can use [KubeKey](https://github.com/kubesphere/kubekey) to install both KubeSphere and
Kubernetes/K3s in single-node mode on your Linux machine. Let's take K3s as an example:

```yaml
# Download KubeKey
curl -sfL https://get-kk.kubesphere.io | VERSION=v3.0.10 sh -
# Make kk executable
chmod +x kk
# Create a cluster
./kk create cluster --with-kubernetes v1.24.14 --container-manager containerd --with-kubesphere v3.4.0
```

You can run the following command to view the installation logs. After KubeSphere is successfully installed, you can
access the KubeSphere web console at `http://IP:30880` and log in using the default administrator account (
admin/P@88w0rd).

```yaml
kubectl logs -n kubesphere-system $(kubectl get pod -n kubesphere-system -l 'app in (ks-install, ks-installer)' -o jsonpath='{.items[0].metadata.name}') -f
``` 

### KubeSphere for hosted Kubernetes services

KubeSphere is hosted on the following cloud providers, and you can try KubeSphere by one-click installation on their
hosted Kubernetes services.

- [KubeSphere for Amazon EKS](https://aws.amazon.com/quickstart/architecture/qingcloud-kubesphere/)
- [KubeSphere for Azure AKS](https://market.azure.cn/marketplace/apps/qingcloud.kubesphere)
- [KubeSphere for DigitalOcean Kubernetes](https://marketplace.digitalocean.com/apps/kubesphere)
- [KubeSphere on QingCloud AppCenter(QKE)](https://www.qingcloud.com/products/kubesphereqke)

You can also install KubeSphere on other hosted Kubernetes services within minutes, see
the [step-by-step guides](https://kubesphere.io/docs/installing-on-kubernetes/) to get started.

> 👨‍💻 No internet access? Refer to
>
the [Air-gapped Installation on Kubernetes](https://kubesphere.io/docs/installing-on-kubernetes/on-prem-kubernetes/install-ks-on-linux-airgapped/)
>
or [Air-gapped Installation on Linux](https://kubesphere.io/docs/installing-on-linux/introduction/air-gapped-installation/)
> for instructions on how to use private registry to install KubeSphere.

## Guidance, discussion, contribution, and support

We :heart: your contribution. The [community](https://github.com/kubesphere/community) walks you through how to get
started contributing KubeSphere.
The [development guide](https://github.com/kubesphere/community/tree/master/developer-guide/development) explains how to
set up development environment.

- [Slack Channel](https://join.slack.com/t/kubesphere/shared_invite/zt-219hq0b5y-el~FMRrJxGM1Egf5vX6QiA)
- [Youtube](https://www.youtube.com/channel/UCyTdUQUYjf7XLjxECx63Hpw)
- [Twitter](https://twitter.com/KubeSphere)

:hugs: Please submit any KubeSphere bugs, issues, and feature requests
to [KubeSphere GitHub Issue](https://github.com/kubesphere/kubesphere/issues).

:heart_decoration: The KubeSphere team also provides efficient official ticket support to respond in hours. For more
information, click [KubeSphere Online Support](https://kubesphere.cloud/en/ticket/).

## Who are using KubeSphere

The [user case studies](https://kubesphere.io/case/) page includes the user list of the project. You
can [leave a comment](https://github.com/kubesphere/kubesphere/issues/4123) to let us know your use case.

## Landscapes

<p align="center">
<br/><br/>
<img src="https://landscape.cncf.io/images/left-logo.svg" width="150"/>&nbsp;&nbsp;<img src="https://landscape.cncf.io/images/right-logo.svg" width="200"/>&nbsp;&nbsp;
<br/><br/>
KubeSphere is a member of CNCF and a <a href="https://www.cncf.io/certification/software-conformance/#logos">Kubernetes Conformance Certified platform
</a>, which enriches the <a href="https://landscape.cncf.io/?landscape=observability-and-analysis&license=apache-license-2-0">CNCF CLOUD NATIVE Landscape.
</a>
</p>
