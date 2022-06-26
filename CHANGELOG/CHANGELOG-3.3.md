- [v3.3.0](#v330)
  - [DevOps](#devops)
  - [Network](#network)
  - [App Store](#app-store)
  - [Multi-tenancy & Multi-cluster](#multi-tenancy--multi-cluster)
  - [Observability](#observability)
  - [Authentication & Authorization](#authentication--authorization)
  - [Storage](#storage)
  - [Service Mesh](#service-mesh)
  - [KubeEdge Integration](#kubeedge-integration)
  - [Development & Testing](#development--testing)
  - [User Experience](#user-experience)
  - [API Changes](#api-changes)



# v3.3.0

## DevOps

### Features

- Add the Continuous Deployment feature, which supports GitOps and uses Argo CD as the backend. ([ks-devops#466](https://github.com/kubesphere/ks-devops/pull/466), [@LinuxSuRen](https://github.com/LinuxSuRen), [console#2990]( https://github.com/kubesphere/console/pull/2990),[@harrisonliu5](https://github.com/harrisonliu5))
- Add a webhook for processing pipeline events. ([ks-devops#442](https://github.com/kubesphere/ks-devops/pull/442), [@JohnNiang](https://github.com/JohnNiang))
- Add support for verifying SCM tokens directly instead of calling a Jenkins API. ([ks-devops#439](https://github.com/kubesphere/ks-devops/pull/439), [@LinuxSuRen](https://github.com/LinuxSuRen))
- Optimize pipeline templates loaded from APIs. ([ks-devops#453](https://github.com/kubesphere/ks-devops/pull/453), [@JohnNiang](https://github.com/JohnNiang), [console#2963](https://github.com/kubesphere/console/pull/2963), [@harrisonliu5](https://github.com/harrisonliu5))
- Add the allowlist feature on the **Basic Information** page of a DevOps project. ([console#3224](https://github.com/kubesphere/console/pull/3224), [@harrisonliu5](https://github.com/harrisonliu5))
- Add the **Events** tab on the pipeline run record page. ([console#3012](https://github.com/kubesphere/console/pull/3012), [@harrisonliu5](https://github.com/harrisonliu5))
- Prevent passing tokens directly to the DevOps SCM verification API to avoid security vulnerabilities. Instead, only secret names are passed to the API. ([console#2943](https://github.com/kubesphere/console/pull/2943), [@mangoGoForward](https://github.com/mangoGoForward))
- Add a message indicating that the S2I and B2I features currently do not support the containerd runtime. ([console#2734](https://github.com/kubesphere/console/pull/2734), [@weili520](https://github.com/weili520))
- Fix an issue in the **Import Code Repository** dialog box, where repositories and organizations are not completely displayed. ([console#3222](https://github.com/kubesphere/console/pull/3222), [@harrisonliu5](https://github.com/harrisonliu5))


## Network

### Enhancements & Updates

- Integrate OpenELB with KubeSphere for exposing LoadBalancer services. ([console#2993](https://github.com/kubesphere/console/pull/2993), [@weili520](https://github.com/weili520))

### Bug Fixes

- Fix an issue where the gateway of a project is not deleted after the project is deleted. ([kubesphere#4626](https://github.com/kubesphere/kubesphere/pull/4626), [@RolandMa1986](https://github.com/RolandMa1986))
- Fix an issue during gateway log query, where the log query connection is not terminated after the query is complete. ([kubesphere#4927](https://github.com/kubesphere/kubesphere/pull/4927), [@qingwave](https://github.com/qingwave))

## App Store

### Bug Fixes

- Fix a ks-controller-manager crash caused by Helm controller NPE errors. ([kubesphere#4602](https://github.com/kubesphere/kubesphere/pull/4602), [@chaunceyjiang](https://github.com/chaunceyjiang))


## Multi-tenancy & Multi-cluster

### Features

- Add a notification to remind users when the kubeconfig certificate of a cluster is about to expire. ([console#3038](https://github.com/kubesphere/console/pull/3038), [@harrisonliu5](https://github.com/harrisonliu5), [kubesphere#4584](https://github.com/kubesphere/kubesphere/pull/4584), [@iawia002](https://github.com/iawia002))
- Add the kubesphere-config configmap, which provides the name of the current cluster. ([kubesphere#4679](https://github.com/kubesphere/kubesphere/pull/4679), [@iawia002](https://github.com/iawia002))
- Add an API for updating cluster kubeconfig information. ([kubesphere#4562](https://github.com/kubesphere/kubesphere/pull/4562), [@iawia002](https://github.com/iawia002))
- Add the cluster member management and cluster role management features. ([console#3061](https://github.com/kubesphere/console/pull/3061), [@harrisonliu5](https://github.com/harrisonliu5))

## Observability

### Enhancements & Upgrades

- Add process and thread monitoring metrics for containers. ([kubesphere#4711](https://github.com/kubesphere/kubesphere/pull/4711), [@junotx](https://github.com/junotx))
- Add disk monitoring metrics that provide usage of each disk. ([kubesphere#4705](https://github.com/kubesphere/kubesphere/pull/4705), [@junotx](https://github.com/junotx))
- Change metric query statements to match Prometheus component upgrades. ([kubesphere#4621](https://github.com/kubesphere/kubesphere/pull/4621), [@junotx](https://github.com/junotx))
- Add support for importing Grafana templates to create custom monitoring dashboards of a namespace scope. ([kubeshere#4446](https://github.com/kubesphere/kubesphere/pull/4446), [@zhu733756](https://github.com/zhu733756), [console#3202](https://github.com/kubesphere/console/pull/3202) [@weili520](https://github.com/weili520))
- Add support for defining separate data retention periods for container logs, resource events, and audit logs. ([ks-installer#1915](https://github.com/kubesphere/ks-installer/pull/1915), [@wenchajun](https://github.com/wenchajun))
- Upgrade Alertmanager from v0.21.0 to v0.23.0.
- Upgrade Grafana from v7.4.3 to v8.3.3.
- Upgrade kube-state-metrics from v1.9.7 to v2.3.0.
- Upgrade node-exporter from v0.18.1 to v1.3.1.
- Upgrade Prometheus from v2.26.0 to v2.34.0.
- Upgrade Prometheus Operator from v0.43.2 to v0.55.1.
- Upgrade kube-rbac-proxy from v0.8.0 to v0.11.0.
- Upgrade configmap-reload from v0.3.0 to v0.5.0.
- Upgrade Thanos from v0.18.0 to v0.25.2.
- Upgrade kube-events from v0.3.0 to v0.4.0.
- Upgrade Fluent Bit Operator from v0.11.0 to v0.13.0.
- Upgrade Fluent Bit from v1.8.3 to v1.8.11.

## Authentication & Authorization

### Features

- Add support for manually disabling and enabling users. ([kubesphere#4695](https://github.com/kubesphere/kubesphere/pull/4695), [@wansir](https://github.com/wansir), [console#3014](https://github.com/kubesphere/console/pull/3014), [@harrisonliu5](https://github.com/harrisonliu5))

## Storage

### Features

- Add the PVC auto expansion feature, which automatically expands PVCs when remaining capacity is insufficient. ([kubesphere#4660](https://github.com/kubesphere/kubesphere/pull/4660), [@f10atin9](https://github.com/f10atin9).[console#3056](https://github.com/kubesphere/console/pull/3056), [@weili520](https://github.com/weili520))
- Add the volume snapshot content management and volume snapshot class management features. ([kubesphere#4596](https://github.com/kubesphere/kubesphere/pull/4596), [@f10atin9](https://github.com/f10atin9), [console#2958](https://github.com/kubesphere/console/pull/2958), [@weili520](https://github.com/weili520), [console#3051](https://github.com/kubesphere/console/pull/3051), [@weili520](https://github.com/weili520))
- Allow users to set authorization rules for storage classes so that storage classes can be used only in specific projects. ([kubesphere#4770](https://github.com/kubesphere/kubesphere/pull/4770), [@f10atin9](https://github.com/f10atin9), [console#3069](https://github.com/kubesphere/console/pull/3069), [@weili520](https://github.com/weili520))
- Provide usage data of each disk. ([console#3063](https://github.com/kubesphere/console/pull/3063), [@harrisonliu5](https://github.com/harrisonliu5))


## Service Mesh

### Bug Fixes

- Resolve port conflicts of virtual services that use multiple protocols. ([kubesphere#4560](https://github.com/kubesphere/kubesphere/pull/4560), [@RolandMa1986](https://github.com/RolandMa1986))


## KubeEdge Integration

### Features
- Add support for logging in to common cluster nodes and edge nodes from the KubeSphere web console. ([kubesphere#4579](https://github.com/kubesphere/kubesphere/pull/4579), [@lynxcat](https://github.com/lynxcat), [console#2888](https://github.com/kubesphere/console/pull/2888), [@lynxcat](https://github.com/lynxcat))

### Enhancements & Upgrades
- Upgrade KubeEdge from v1.7.2 to v1.9.2.
- Remove EdgeWatcher as KubeEdge v1.9.2 provides similar functions.

## Development & Testing

### Features

- Add the --controllers flag to ks-controller-manager, which allows users to select controllers to be enabled for resource usage reduction during debugging. ([kubesphere#4512](https://github.com/kubesphere/kubesphere/pull/4512), [@live77](https://github.com/live77))
- Prevent ks-apiserver and ks-controller-manager from restarting when the cluster configuration is changed. ([kubesphere#4659](https://github.com/kubesphere/kubesphere/pull/4659), [@x893675](https://github.com/x893675))
- Add an agent to report additional information about ks-apiserver and controller-manager in debugging mode. ([kubesphere#4928](https://github.com/kubesphere/kubesphere/pull/4928), [@xyz-li](https://github.com/xyz-li))


## User Experience

- Add support for more languages on the KubeSphere web console. ([console#2782](https://github.com/kubesphere/console/pull/2782), [@xuliwenwenwen](https://github.com/xuliwenwenwen))
- Add the lifecycle management feature for containers. ([console#2940](https://github.com/kubesphere/console/pull/2940), [@harrisonliu5](https://github.com/harrisonliu5))
- Add support for creating container environment variables in batches from secrets and configmaps. ([console#3044](https://github.com/kubesphere/console/pull/3044), [@weili520](https://github.com/weili520))
- Add a message in the **Audit Log Search** dialog box, which prompts users to enable the audit logs feature. ([console#3062](https://github.com/kubesphere/console/pull/3062), [@harrisonliu5](https://github.com/harrisonliu5))
- Add data units in the **Create Storage Class** dialog box. ([console#3200](https://github.com/kubesphere/console/pull/3200), [@weili520](https://github.com/weili520))
- Add the cluster viewing permission to a user when the user adds a cluster. ([console#3296](https://github.com/kubesphere/console/pull/3296), [@harrisonliu5](https://github.com/harrisonliu5))
- Optimize the service details area on the **Service Topology** page. ([console#2945](https://github.com/kubesphere/console/pull/2945), [@tracer1023](https://github.com/tracer1023))
- Prevent passwords without uppercase letters set through the backend CLI. ([kubesphere#4481](https://github.com/kubesphere/kubesphere/pull/4481), [@live77](https://github.com/live77))
- Fix an issue where statefulset creation fails when a volume is mounted to an init container. ([console#2730](https://github.com/kubesphere/console/pull/2730), [@weili520](https://github.com/weili520))
- Fix an app installation failure, which occurs when users click buttons too fast. ([console#2735](https://github.com/kubesphere/console/pull/2735), [@weili520](https://github.com/weili520))
- Change the abbreviation of "ReadOnlyMany" from "ROM" to "ROX". ([console#2751](https://github.com/kubesphere/console/pull/2751), [@123liubao](https://github.com/123liubao))
- Fix an incorrect message displayed during workload creation, which indicates that resource requests have exceeded limits. ([console#2809](https://github.com/kubesphere/console/pull/2809) , [@weili520](https://github.com/weili520))
- Fix an issue where no data is displayed on the **Traffic Management** and **Tracing** tab pages in a multi-cluster project. ([console#3195](https://github.com/kubesphere/console/pull/3195), [@harrisonliu5](https://github.com/harrisonliu5))
- Set the **Token** parameter on the webhook settings page as mandatory. ([console#2903](https://github.com/kubesphere/console/pull/2903), [@xuliwenwenwen](https://github.com/xuliwenwenwen))
- Improve user experience on the web console of member clusters. ([kubesphere#4721](https://github.com/kubesphere/kubesphere/pull/4721), [@iawia002](https://github.com/iawia002), [console#3031](https://github.com/kubesphere/console/pull/3031), [@weili520](https://github.com/weili520))
- Optimize the UI text of cluster, workspace, and project deletion. ([console#3004](https://github.com/kubesphere/console/pull/3004), [@weili520](https://github.com/weili520))
- Fix an issue where data in the service details area on the **Service Topology** page is not updated automatically. ([console#3117](https://github.com/kubesphere/console/pull/3117),[@weili520](https://github.com/weili520))
- Fix an issue where existing settings are not displayed when a stateful service is edited. ([console#2845](https://github.com/kubesphere/console/pull/2845), [@weili520](https://github.com/weili520))
- Add a time range selector on the **Traffic Monitoring** tab page. ([console#2916]( https://github.com/kubesphere/console/pull/2916), [@weili520](https://github.com/weili520))
- Optimize the **Service Type** and **External Access** columns of the service list. ([console#3058](https://github.com/kubesphere/console/pull/3058),[@weili520](https://github.com/weili520))
- Fix an issue where container probes are still displayed after they are deleted. ([console#3213](https://github.com/kubesphere/console/pull/3213), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix incorrect container names in the service details area on the **Service Topology** page. ([console#3276](https://github.com/kubesphere/console/pull/3276), [@weili520](https://github.com/weili520))
- Fix the incorrect number of worker nodes on the **Cluster Nodes** page. ([console#3279](https://github.com/kubesphere/console/pull/3279), [@weili520](https://github.com/weili520))
- Fix a workspace API error caused by an incorrect cluster role. ([console#3297](https://github.com/kubesphere/console/pull/3297), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix an incorrect error message displayed when kubeconfig is updated for a second time. ([console#3392](https://github.com/kubesphere/console/pull/3392), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix goroutine leaks that occur when a container terminal is opened. ([kubesphere#4918](https://github.com/kubesphere/kubesphere/pull/4918), [@anhoder](https://github.com/anhoder))
- Fix unexpected ks-apiserver panic caused by an index out of range error of metrics API calling. ([kubesphere#4691](https://github.com/kubesphere/kubesphere/pull/4691), [@larryliuqing](https://github.com/larryliuqing))
- Fix an ks-apiserver crash caused by resource discovery failure. ([kubesphere#4835](https://github.com/kubesphere/kubesphere/pull/4835),  [@wansir](https://github.com/wansir))
- Fix an incorrect private key type in kubeconfig information. ([kubesphere#4936](https://github.com/kubesphere/kubesphere/pull/4936), [@xyz-li](https://github.com/xyz-li))
- Improve the registry verification API to resolve registry verification failures. ([kubesphere#4678](https://github.com/kubesphere/kubesphere/pull/4678), [@wansir](https://github.com/wansir))

## API Changes

- Expose kube-apiserver of the host cluster as a LoadBalancer service for member clusters to access. ([kubesphere#4528](https://github.com/kubesphere/kubesphere/pull/4528), [@lxm](https://github.com/lxm))
- Provide RESTful APIs for ClusterTemplate. ([ks-devops#468](https://github.com/kubesphere/ks-devops/pull/468), [@JohnNiang](https://github.com/JohnNiang))
- Provide template-related APIs. ([ks-devops#460](https://github.com/kubesphere/ks-devops/pull/460), [@JohnNiang](https://github.com/JohnNiang))
- Change the KubeEdge proxy service to `http://edgeservice.kubeedge.svc/api/`. ([kubesphere#4478](https://github.com/kubesphere/kubesphere/pull/4478), [@zhu733756](https://github.com/zhu733756))




