KubeSphere Roadmap demonstrates a list of open source product development plans and features being split by the edition and modules, as well as KubeSphere community's anticipation. Obviously, it details the future's direction of KubeSphere, but may change over time. We hope that can help you to get familiar with the project plans and vision through the Roadmap. Of course, if you have any better ideas, welcome to filing [Issues](https://github.com/kubesphere/kubesphere/issues).

# Release Goals

| Edition  | Schedule |
|---|---|
| Release Express| Jul, 2018 |
| Release v1.0.0| Dec, 2018 |
| Release v1.0.1| Jan, 2019 |
| Release v2.0.0| May, 2019 |
| Release v2.0.1| June, 2019|
| Release v2.0.2| Jul, 2019 |
| Release v2.1.0| Nov, 2019 |
| Release v2.1.1| Feb, 2020 |
| Release v3.0.0| June, 2020 |

# v3.1
## **Feature:**

### IAM:

- [x] Add account confirmation process, solve problems caused by special characters in username.[#2953](https://github.com/kubesphere/kubesphere/issues/2953)
- [x] Support [CAS](https://apereo.github.io/cas/5.0.x/protocol/CAS-Protocol-Specification.html) identity provider.[#3047](https://github.com/kubesphere/kubesphere/issues/3047)
- [x] Support [OIDC](https://openid.net/specs/openid-connect-core-1_0.html) identity provider.[#2941](https://github.com/kubesphere/kubesphere/issues/2941)
- [x] Support IDaaS(Alibaba Cloud Identity as a Service) identity provider.[#2997](https://github.com/kubesphere/kubesphere/pull/2997)
- [x] Improve LDAP identity provider, support LDAPS and search filter.[#2970](https://github.com/kubesphere/kubesphere/issues/2970)
- [x] Improve identity provider plugin, simplify identity provider configuration.[#2970](https://github.com/kubesphere/kubesphere/issues/2970)
- [x] Limit login record maximum entries.[#3191](https://github.com/kubesphere/kubesphere/issues/3191)
- [x] Service account management, allows to assign role to Service Account [#3211](https://github.com/kubesphere/kubesphere/issues/3211)

### Multitenancy:

- [ ] Support transfer namespace to another workspace.[#3028](https://github.com/kubesphere/kubesphere/issues/3028)
- [x] Add user group, now users can be assigned to a group and invite a group to a workspace or a project.[#2940](https://github.com/kubesphere/kubesphere/issues/2940)
- [x] Add resource quota to a workspace. The resource quota is the same with Kubernetes[ResourceQuota](https://kubernetes.io/docs/concepts/policy/resource-quotas/), providesconstraints that limit aggregate resource consumption by all namespaces within.[  ](https://kubernetes.io/docs/concepts/policy/resource-quotas/)[#2939](https://github.com/kubesphere/kubesphere/issues/2939)
- [x] Add choose of whether to cascade delete related resources when deleting workspace.[#3192](https://github.com/kubesphere/kubesphere/issues/3192)

### DevOps:

- [x] Support Gitlab with KubeSphere multi-branch pipeline. [#3100](https://github.com/kubesphere/kubesphere/issues/3100)
- [x] Run multiple DevOps pipelines at the same time instead of just one.[#1811](https://github.com/kubesphere/kubesphere/issues/1811)
- [x] Create a Pipeline from an existing one. Users now can create exactly same pipeline from an existing one.[#3053](https://github.com/kubesphere/kubesphere/issues/3053)
- [x] Add approval control for pipelines. [#2483](https://github.com/kubesphere/kubesphere/issues/2483) [#3006](https://github.com/kubesphere/kubesphere/issues/3006)
- [x] Add and display the status of the pipeline on the front page. [#3007](https://github.com/kubesphere/kubesphere/issues/3007)
- [x] Support tag trigger pipeline. [#3051](https://github.com/kubesphere/kubesphere/issues/3051)
- [x] Using human-readable error message for pipeline cron text. [#2919](https://github.com/kubesphere/kubesphere/issues/2919)
- [x] Interactive creation pipeline. [#1283](https://github.com/kubesphere/console/issues/1283)
- [x] Add S2I webhook support. [#6](https://github.com/kubesphere/s2ioperator/issues/6)
- [x] Using human-readable error message for S2I. [#140](https://github.com/kubesphere/s2ioperator/issues/140)

### microservice

- [x] Support graph traffic direction. It will show outbound/inbound traffic direction in composed application traffic graph. [#3153](https://github.com/kubesphere/kubesphere/issues/3153)
- [x] Support Kiali addon. Users can manage istio by kiali directly. [#3106](https://github.com/kubesphere/kubesphere/issues/3106)
- [x] Support Nginx Ingress Gateway monitor. Add prometheus metrics for nginx ingress controller. [#1205](https://github.com/kubesphere/ks-installer/pull/1205)
- [x] Support add router component. [#1426](https://github.com/kubesphere/console/issues/1426) 

### KubeEdge Integration [#3070](https://github.com/kubesphere/kubesphere/issues/3070)

- [x] KubeEdge cloud components setup.
- [x] KubeEdge edge nodes setup.
- [x] Edge nodes logging and metrics support.
- [x] Automatic network configuration on edge node joining/leaving.
- [x] Automatic taint edge node on joining.
- [x] Cloud workloads (such as daemonset) with wide tolerations should not be scheduled to edge node by adding a node selector.
- [x] Scheduling workloads to edge nodes.

### Observability

- [x] Utilizing existing Prometheus stack setup. [#3068](https://github.com/kubesphere/kubesphere/issues/3068) [#1164](https://github.com/kubesphere/ks-installer/pull/1164) [Guide](https://kubesphere.io/docs/faq/observability/byop/)

#### Custom monitoring [#3067](https://github.com/kubesphere/kubesphere/issues/3067)

- [x] Configure ServiceMonitor via UI. [#1031](https://github.com/kubesphere/console/pull/1301) 
- [x] PromQL auto-completion and syntax highlighting. [#1307](https://github.com/kubesphere/console/pull/1307)
- [x] Support cluster-level custom monitoring. [#3193](https://github.com/kubesphere/kubesphere/pull/3193)
- [x] Tools to convert Grafana dashboard to KubeSphere Dashboard [#9](https://github.com/kubesphere/monitoring-dashboard/pull/9)

#### Custom Alerting [#3065](https://github.com/kubesphere/kubesphere/issues/3065)

- [x] Prometheus alert rule management. [#3181](https://github.com/kubesphere/kubesphere/pull/3181)
- [x] Alert rule tenant control: global/namespace level alert rules. [#3181](https://github.com/kubesphere/kubesphere/pull/3181)
- [x] List alerts for a specific alert rule. [#3181](https://github.com/kubesphere/kubesphere/pull/3181)

#### Multi-tenant Notification support including Email/DingTalk/Slack/Wechat works/Webhook [#3066](https://github.com/kubesphere/kubesphere/issues/3066)

- [x] More notification channels including Email, DingTalk, Slack, WeCom, Webhook
- [ ] Multi-tenant control of notification

#### Logging

- [x] Support output logs to [Loki](https://github.com/kubesphere/fluentbit-operator/blob/master/docs/plugins/output/loki.md). [#39](https://github.com/kubesphere/fluentbit-operator/pull/39)
- [x] Support collecting kubelet/docker/containerd logs. [#38](https://github.com/kubesphere/fluentbit-operator/pull/38)
- [x] Support collecting [auditd](https://github.com/kubesphere/fluentbit-operator#auditd) log. [#45](https://github.com/kubesphere/fluentbit-operator/pull/45)
- [ ] Support containerd log format

### Application Lifecycle Management (OpenPitrix)

- [x] Refactoring OpenPitrix with CRD, while fix bugs caused by legacy architecture [#3036](https://github.com/kubesphere/kubesphere/issues/3036) [#3001](https://github.com/kubesphere/kubesphere/issues/3001) [#2995](https://github.com/kubesphere/kubesphere/issues/2995) [#2981](https://github.com/kubesphere/kubesphere/issues/2981) [#2954](https://github.com/kubesphere/kubesphere/issues/2954) [#2951](https://github.com/kubesphere/kubesphere/issues/2951) [#2783](https://github.com/kubesphere/kubesphere/issues/2783) [#2713](https://github.com/kubesphere/kubesphere/issues/2713) [#2700](https://github.com/kubesphere/kubesphere/issues/2700) [#1903](https://github.com/kubesphere/kubesphere/issues/1903) 

### Network

- [x] IPPool for Calico and VMs [#3057](https://github.com/kubesphere/kubesphere/issues/3057)
- [x] Support for deployment using static IPs [#3058](https://github.com/kubesphere/kubesphere/issues/3058)
- [x] Support network visualization [#3061](https://github.com/kubesphere/kubesphere/issues/3061) [#583](https://github.com/kubesphere/kubesphere/issues/583)

### Metering

- [x] Support for viewing resource consumption at the cluster, workspace, and application template levels [#3062](https://github.com/kubesphere/kubesphere/issues/3062)

### MultiCluster:

- [x] Validate member cluster configuration if the member cluster config(e.g. jwtSecret) is not same as host cluster, which can reduce the complexity of joining clusters. [#3232](https://github.com/kubesphere/kubesphere/issues/3232)
- [x] Support configuring cluster controller resync period, some users need to update the cluster component more quickly. [#3213](https://github.com/kubesphere/kubesphere/issues/3213)
- [x] Lightweight member cluster installation, support running kubepshere without redis and ldap and so on. [#3056](https://github.com/kubesphere/kubesphere/issues/3056)
- [x] Refactor cluster controller(it should updates the cluster.staus field instead of the tower server when using proxy connection).  [#3234](https://github.com/kubesphere/kubesphere/issues/3234)
- [ ] Support the tower server and agent running highly available(server with leader election option and running more than one agent). [#31](https://github.com/kubesphere/tower/issues/31)

## **Upgrade:**

- [x] Upgrade isito version from 1.4.8 => 1.6.10[#3326](https://github.com/kubesphere/kubesphere/issues/3236)
- [x] Upgrade prometheus client_golang to v1.5.1, upgrade prometheus to v1.8.2 [3097](https://github.com/kubesphere/kubesphere/pull/3097)
- [x] Upgrade Kubectl version for the Toolbox, and the Kubectl Verion will be matched with the Kubernetes Server version. [#3103](https://github.com/kubesphere/kubesphere/issues/3103)
- [x] Upgrade dependencies of K8s version from v1.17 to V1.18. [#3274](https://github.com/kubesphere/kubesphere/issues/3274)

### DevOps:

- [x] Upgrade Jenkins Version to 2.249.1. [#2618](https://github.com/kubesphere/kubesphere/issues/2618)
- [x] Using Jenkins distribution solution to deploy Jenkins. [#2182](https://github.com/kubesphere/kubesphere/issues/2182)

### Observability

- [x] Upgrade Prometheus to v2.26.0
- [x] Upgrade kube-state-metrics to v1.9.7
- [x] Upgrade Notification Manager to v1.0.0 [Releases](https://github.com/kubesphere/notification-manager/releases)
- [x] Upgrade FluentBit Operator to v0.5.0 [Releases](https://github.com/kubesphere/fluentbit-operator/releases)
- [x] Upgrade FluentBit to v1.6.9

## **BugFix:**

- [x] Fix unable to get service mesh graph when in a namespace whose name starts with kube[#3126](https://github.com/kubesphere/kubesphere/issues/3162)
- [x] Fix deploy workloads to kubernetes encountered bad_certificate error. [#3112](https://github.com/kubesphere/kubesphere/issues/3112)
- [x] Fix DevOps project admin cannot download artifacts. [#3088](https://github.com/kubesphere/kubesphere/issues/3083)
- [x] Fix cannot create pipeline caused by user admin not found in directory. [#3105](https://github.com/kubesphere/kubesphere/issues/3105)
- [x] Fix the security risk caused by pod viewer can connect to the container terminal. [#3041](https://github.com/kubesphere/kubesphere/issues/3041)
- [ ] Fix some resources cannot be deleted in cascade. [#2912](https://github.com/kubesphere/kubesphere/issues/2912)
- [x] Fix self-signed certificate for admission webhook relies on legacy Common Name field. [#2928](https://github.com/kubesphere/kubesphere/issues/2928)
- [x] Fix application monitor view [#1394](https://github.com/kubesphere/console/issues/1394)
- [x] Fix Grayscale Release when service app label is not the same with the servicename [#3128](https://github.com/kubesphere/kubesphere/issues/3128)
- [x] Fix application status which is always updating [#3241](https://github.com/kubesphere/kubesphere/issues/3241)
- [x] Fix workspaces on member cluster would be deleted when joining to host if there are workspaces with same name on the host. [#3169](https://github.com/kubesphere/kubesphere/issues/3169)
- [x] Fix cluster unjoin federation plane when using proxy connection. [#3202](https://github.com/kubesphere/kubesphere/pull/3203)
- [x] Fix multi-cluster-status-condition. [#3135](https://github.com/kubesphere/kubesphere/issues/3135)
- [x] Fix edge cluster can't join to the host cluster. [#3198](https://github.com/kubesphere/kubesphere/issues/3198)
- [x] Fix can not rollout restart tower agent. [#29](https://github.com/kubesphere/tower/issues/29)
- [x] Fix failed calling webhook due to certificate relies on legacy Common Name field. [#2928](https://github.com/kubesphere/kubesphere/issues/2928)

### DevOps:

- [x] Fix error webhook link under multicluster. [forum #2626](https://kubesphere.com.cn/forum/d/2626-webhook-jenkins)
- [x] Fix some data lost in jenkinsfile when users edit pipeline. [#1270](https://github.com/kubesphere/console/issues/1270)
- [x] Fix error when clicking "Docker Container Registry Credentials". [console #1269](https://github.com/kubesphere/console/issues/1269)
- [x] Fix chinese show in code quality under English language. [consoel #1278](https://github.com/kubesphere/console/issues/1278)
- [x] Fix get the error when it displays a bool parameter from InSCM Jenkinsfile. [#3043](https://github.com/kubesphere/kubesphere/issues/3043)

# v3.0

## Multi-cluster

## DevOps

- [x] Create / Edit Pipeline Process Optimization.
- [x] S2I/B2I supports webhook.
- [x] Image registry optimization.
- [x] Pipeline support integration with JIRA.
- [x] Pipeline integrates the notification of KubeSphere.
- [x] Pipeline integrates KubeSphere custom monitoring.

## Observability

- [ ] Logging console enhancement
- [ ] Monitoring stack upgrade including Prometheus, Prometheus Operator, Node exporter, kube-state-metrics etc.
- [ ] Custom metrics support including application custom metrics dashboard, custom metrics HPA
- [ ] Integration with Alertmanager
- [ ] K8s Event management
- [ ] K8s Audit Support
- [ ] Notification Enhancement

## Network

## Storage

- [ ] Snapshot management
- [ ] Volume cloning
- [ ] Volume monitoring and alerting
- [ ] Identify storage system capabilities
- [ ] Restore volume to available status
- [ ] Unified integrate third-party storage plugin

## Security & Multitenancy

- [ ] Support the OAuth2 SSO plugin.
- [ ] Workspace resource quota.
- [ ] Refactor access management framework to adapt to multi-cluster design.

## Application Lifecycle Management (OpenPitrix)


# v2.1

- [ ] Most of the work will be bugfix
- [ ] Refactor RBAC in order to support future versions regarding third-party plugins with custom access control.
- [ ] Refactor installer
- [ ] FluentBit Operator upgrade
