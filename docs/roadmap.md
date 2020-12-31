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

### Multitenancy:

- [ ] Add user group, now users can be assigned to a group and invite a group to a workspace or a project.[#2940](https://github.com/kubesphere/kubesphere/issues/2940)
- [ ] Add resource quota to a workspace. The resource quota is the same with Kubernetes[ResourceQuota](https://kubernetes.io/docs/concepts/policy/resource-quotas/), providesconstraints that limit aggregate resource consumption by all namespaces within.[  ](https://kubernetes.io/docs/concepts/policy/resource-quotas/)[#2939](https://github.com/kubesphere/kubesphere/issues/2939)
- [ ] Add choose of whether to cascade delete related resources when deleting workspace.[#3192](https://github.com/kubesphere/kubesphere/issues/3192)
### DevOps:

- [x] Run multiple DevOps pipelines at the same time instead of just one.[#1811](https://github.com/kubesphere/kubesphere/issues/1811)
- [x] Clone pipeline. Users now can create exactly same pipeline from an existing one.[#3053](https://github.com/kubesphere/kubesphere/issues/3053)

### IAM:
- [ ] Service account management, allows to assign role to Service Account [#3211](https://github.com/kubesphere/kubesphere/issues/3211)

## **Upgrade:**

- [ ] Upgrade isito version from 1.4.8 => 1.6.5[#3326](https://github.com/kubesphere/kubesphere/issues/3236)

- [x] Upgrade Kubectl version for the Toolbox, and the Kubectl Verion will be matched with the Kubernetes Server version. [#3103](https://github.com/kubesphere/kubesphere/issues/3103)

## **BugFix:**

- [x] Fix unable to get service mesh graph when in a namespace whose name starts with kube[#3126](https://github.com/kubesphere/kubesphere/issues/3162)
- [x] Fix workspaces on member cluster would be deleted when joining to host if there are workspaces with same name on the host. [#3169](https://github.com/kubesphere/kubesphere/issues/3169)


# v3.0

## Multi-cluster

## DevOps

- [ ] Create / Edit Pipeline Process Optimization.
- [ ] S2I/B2I supports webhook.
- [ ] Image registry optimization.
- [ ] Pipeline support integration with JIRA.
- [ ] Pipeline integrates the notification of KubeSphere.
- [ ] Pipeline integrates KubeSphere custom monitoring.

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
