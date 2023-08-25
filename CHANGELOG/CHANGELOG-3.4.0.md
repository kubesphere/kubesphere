- [v3.4.0](#v340)
  - [DevOps](#devops)
    - [Enhancements \& Updates](#enhancements--updates)
    - [Bug Fixes](#bug-fixes)
  - [Storage](#storage)
    - [Bug Fixes](#bug-fixes-1)
  - [Gateway and Microservice](#gateway-and-microservice)
    - [Features](#features)
    - [Enhancements \& Updates](#enhancements--updates-1)
    - [Bug Fixes](#bug-fixes-2)
  - [Observability](#observability)
    - [Features](#features-1)
    - [Enhancements \& Updates](#enhancements--updates-2)
    - [Bug Fixes](#bug-fixes-3)
  - [Multi-tenancy and Multi-cluster](#multi-tenancy-and-multi-cluster)
    - [Enhancements \& Updates](#enhancements--updates-3)
    - [Bug Fixes](#bug-fixes-4)
  - [App Store](#app-store)
    - [Bug Fixes](#bug-fixes-5)
  - [Network](#network)
    - [Enhancements \& Updates](#enhancements--updates-4)
  - [Authentication and Authorization](#authentication-and-authorization)
    - [Features](#features-2)
    - [Enhancements \& Updates](#enhancements--updates-5)
    - [Bug Fixes](#bug-fixes-6)
  - [API Changes](#api-changes)
  - [User Experience](#user-experience)
    - [Features](#features-3)
    - [Bug Fixes](#bug-fixes-7)
    - [Enhancements \& Updates](#enhancements--updates-6)


# v3.4.0

## DevOps

### Enhancements & Updates

- Support user-defined pipeline configuration steps. 
([ks-devops#768](https://github.com/kubesphere/ks-devops/pull/768), [@LinuxSuRen](https://github.com/LinuxSuRen))
- Optimize the devops-jenkins JVM memory configuration. 
([ks-installer#2206](https://github.com/kubesphere/ks-installer/pull/2206), [@yudong2015](https://github.com/yudong2015))

### Bug Fixes

- Fix the issue of removing ArgoCD resources without cascade parameters. ([ks-devops#949](https://github.com/kubesphere/ks-devops/pull/949), [@chilianyi](https://github.com/chilianyi))
- Fix the issue that downloading artifacts for multi-branch pipelines fails. 
([ks-devops#969](https://github.com/kubesphere/ks-devops/pull/969), [@littlejiancc](https://github.com/littlejiancc))
- Fix the issue that the pipeline running status is inconsistent with Jenkins (Add retry for pipelinerun annotation update). 
([ks-devops#907](https://github.com/kubesphere/ks-devops/pull/907), [@yudong2015](https://github.com/yudong2015))
- Fix the issue that the running of a pipeline created by a new user is pending. 
([ks-devops#936](https://github.com/kubesphere/ks-devops/pull/936), [@yudong2015](https://github.com/yudong2015))


## Storage

### Bug Fixes

- Fix the issue that pvc cannot be deleted.
([kubesphere#5271](https://github.com/kubesphere/kubesphere/pull/5271), [@f10atin9](https://github.com/f10atin9))

## Gateway and Microservice

### Features

- Gateway supports the configuration of forwarding TCP/UDP traffic.([kubesphere#5445](https://github.com/kubesphere/kubesphere/pull/5445), [@hongzhouzi](https://github.com/hongzhouzi))

### Enhancements & Updates

- Upgrade ingress nginx: v1.1.0 -> v1.3.1.
([kubesphere#5490](https://github.com/kubesphere/kubesphere/pull/5490), [@hongzhouzi](https://github.com/hongzhouzi))
- Upgrade servicemesh: 
istio: 1.11.1 -> 1.14.6; kiali: v1.38.1 -> v1.50.1; jaeger: 1.27 -> 1.29.([kubesphere#5792](https://github.com/kubesphere/kubesphere/pull/5792), [@hongzhouzi](https://github.com/hongzhouzi))

### Bug Fixes

- Fix the issue that the returned cluster gateways duplicate. ([kubesphere#5582](https://github.com/kubesphere/kubesphere/pull/5582), [@hongzhouzi](https://github.com/hongzhouzi))
- Fix the verification error when upgrading the gateway. 
([kubesphere#5232](https://github.com/kubesphere/kubesphere/pull/5232), [@hongzhouzi](https://github.com/hongzhouzi))
- Fix the abnormal display of cluster gateway log and resource status after changing gateway namespace configuration. 
([kubesphere#5248](https://github.com/kubesphere/kubesphere/pull/5248), [@hongzhouzi](https://github.com/hongzhouzi))


## Observability

### Features

- Add CRDs such as RuleGroup, ClusterRuleGroup, GlobalRuleGroup to support Alerting v2beta1 APIs. 
([kubesphere#5064](https://github.com/kubesphere/kubesphere/pull/5064), [@junotx](https://github.com/junotx))
- Add admission webhook for RuleGroup, ClusterRuleGroup, GlobalRuleGroup. 
([kubesphere#5071](https://github.com/kubesphere/kubesphere/pull/5071), [@junotx](https://github.com/junotx))
- Add controllers to sync RuleGroup, ClusterRuleGroup, GlobalRuleGroup resources to PrometheusRule resources. 
([kubesphere#5081](https://github.com/kubesphere/kubesphere/pull/5081), [@junotx](https://github.com/junotx))
- Add Alerting v2beta1 APIs. 
([kubesphere#5115](https://github.com/kubesphere/kubesphere/pull/5115), [@junotx](https://github.com/junotx))
- The ks-apiserver of Kubesphere integrates the v1 and v2 versions of opensearch, and users can use the external or built-in opensearch cluster for log storage and query. (Currently the built-in opensearch version of Kubesphere is v2). 
([kubesphere#5044](https://github.com/kubesphere/kubesphere/pull/5044), [@wenchajun](https://github.com/wenchajun))
- ks-installer integrates the opensearch dashboard, which should be enabled by users. 
([ks-installer#2197](https://github.com/kubesphere/ks-installer/pull/2197), [@wenchajun](https://github.com/wenchajun))

### Enhancements & Updates
- Upgrade Prometheus stack dependencies. 
([kubesphere#5520](https://github.com/kubesphere/kubesphere/pull/5520), [@junotx](https://github.com/junotx))
- Support configuring the maximum number of logs that can be exported. ([kubesphere#5794](https://github.com/kubesphere/kubesphere/pull/5794), [@wansir](https://github.com/wansir))
- The monitoring component supports Kubernetes PDB Apiversion changes.  ([ks-installer#2190](https://github.com/kubesphere/ks-installer/pull/2190), [@frezes](https://github.com/frezes))
- Upgrade Notification Manager to v2.3.0. 
([kubesphere#5030](https://github.com/kubesphere/kubesphere/pull/5030), [@wanjunlei](https://github.com/wanjunlei))
- Support cleaning up notification configuration in member clusters when a member cluster is deleted. 
([kubesphere#5077](https://github.com/kubesphere/kubesphere/pull/5077), [@wanjunlei](https://github.com/wanjunlei))
- Support switching notification languages. 
([kubesphere#5088](https://github.com/kubesphere/kubesphere/pull/5088), [@wanjunlei](https://github.com/wanjunlei))
- Support route notifications to specified users. 
([kubesphere#5206](https://github.com/kubesphere/kubesphere/pull/5206), [@wanjunlei](https://github.com/wanjunlei))

### Bug Fixes

- Fix the issue that Goroutine leaks when getting audit event sender times out. ([kubesphere#5342](https://github.com/kubesphere/kubesphere/pull/5342), [@hzhhong](https://github.com/hzhhong))
- Fix the promql statement of ingress P95 delay. 
([kubesphere#5119](https://github.com/kubesphere/kubesphere/pull/5119), [@iawia002](https://github.com/iawia002))


## Multi-tenancy and Multi-cluster

### Enhancements & Updates

- Check the cluster ID (kube-system UID) when updating the cluster.  ([kubesphere#5299](https://github.com/kubesphere/kubesphere/pull/5299), [@yzxiu](https://github.com/yzxiu))

### Bug Fixes

- Make sure the cluster is Ready when cleaning up notifications.([kubesphere#5392](https://github.com/kubesphere/kubesphere/pull/5392), [@iawia002](https://github.com/iawia002))
- Fix the webhook validation issue for new clusters. 
([kubesphere#5802](https://github.com/kubesphere/kubesphere/pull/5802), [@iawia002](https://github.com/iawia002))
- Fix the incorrect cluster status. 
([kubesphere#5130](https://github.com/kubesphere/kubesphere/pull/5130), [@x893675](https://github.com/x893675))
- Fix the issue of potentially duplicated entries for granted clusters in the workspace.
([kubesphere#5795](https://github.com/kubesphere/kubesphere/pull/5795), [@wansir](https://github.com/wansir))


## App Store

### Bug Fixes

- Fix the ID generation failure in IPv6-only environment. 
([kubesphere#5419](https://github.com/kubesphere/kubesphere/pull/5419), [@isyes](https://github.com/isyes))
- Fix the missing Home field in app templates. 
([kubesphere#5425](https://github.com/kubesphere/kubesphere/pull/5425), [@liangzai006](https://github.com/liangzai006))
- Fix the issue that the uploaded app templates do not show icons. ([kubesphere#5467](https://github.com/kubesphere/kubesphere/pull/5467), [@liangzai006](https://github.com/liangzai006))
- Fix missing maintainers in Helm apps. 
([kubesphere#5401](https://github.com/kubesphere/kubesphere/pull/5401), [@qingwave](https://github.com/qingwave))
- Fix the issue that Helm applications in a failed status cannot be upgraded again. 
([kubesphere#5543](https://github.com/kubesphere/kubesphere/pull/5543), [@sekfung](https://github.com/sekfung))
- Fix the wrong "applicationId" parameter. 
([kubesphere#5666](https://github.com/kubesphere/kubesphere/pull/5666), [@sologgfun](https://github.com/sologgfun))
- Fix the infinite loop after app installation failure. 
([kubesphere#5793](https://github.com/kubesphere/kubesphere/pull/5793), [@wansir](https://github.com/wansir))
- FIx the wrong status of application repository. 
([kubesphere#5152](https://github.com/kubesphere/kubesphere/pull/5152), [@x893675](https://github.com/x893675))


## Network

### Enhancements & Updates

- Upgrade dependencies. ([kubesphere#5557](https://github.com/kubesphere/kubesphere/pull/5557), [@renyunkang](https://github.com/renyunkang))


## Authentication and Authorization

### Features

- Add inmemory cache. ([kubesphere#4894](https://github.com/kubesphere/kubesphere/pull/4894), [@zhou1203](https://github.com/zhou1203))
- Add Resource Getter v1beta1. ([kubesphere#5416](https://github.com/kubesphere/kubesphere/pull/5416), [@zhou1203](https://github.com/zhou1203))
- Add write operation for Resource Manager. ([kubesphere#5601](https://github.com/kubesphere/kubesphere/pull/5601), [@zhou1203](https://github.com/zhou1203))

### Enhancements & Updates

- Add iam.kubesphere/v1beta1 RoleTemplate. ([kubesphere#5080](https://github.com/kubesphere/kubesphere/pull/5080), [@zhou1203](https://github.com/zhou1203))
- Update the password minimum length to 8. ([kubesphere#5516](https://github.com/kubesphere/kubesphere/pull/5516), [@zhou1203](https://github.com/zhou1203))
- Update Version API. ([kubesphere#5542](https://github.com/kubesphere/kubesphere/pull/5542), [@zhou1203](https://github.com/zhou1203))
- Update identityProvider API. 
([kubesphere#5534](https://github.com/kubesphere/kubesphere/pull/5534), [@zhou1203](https://github.com/zhou1203))
- Add IAM v1beta1 APIs. 
([kubesphere#5502](https://github.com/kubesphere/kubesphere/pull/5502), [@zhou1203](https://github.com/zhou1203))

### Bug Fixes

- Fix the issue that the enableMultiLogin configuration does not take effect. ([kubesphere#5819](https://github.com/kubesphere/kubesphere/pull/5819), [@wansir](https://github.com/wansir))

## API Changes

- Use autoscaling/v2 API. ([kubesphere#5833](https://github.com/kubesphere/kubesphere/pull/5833), [@LQBing](https://github.com/LQBing))
- Use batch/v1 API. ([kubesphere#5562](https://github.com/kubesphere/kubesphere/pull/5562), [@wansir](https://github.com/wansir))
- Update health check API. ([kubesphere#5496](https://github.com/kubesphere/kubesphere/pull/5496), [@smartcat99](https://github.com/smartcat99))
- Fix the ks-apiserver crash issue in K8s v1.25. ([kubesphere#5428](https://github.com/kubesphere/kubesphere/pull/5428), [@smartcat999](https://github.com/smartcat999))


## User Experience

### Features

- Resource API supports searching alias in annotations. ([kubesphere#5807](https://github.com/kubesphere/kubesphere/pull/5807), [@iawia002](https://github.com/iawia002))

### Bug Fixes

- Fix the potential Websocket link leakage issue. ([kubesphere#5024](https://github.com/kubesphere/kubesphere/pull/5024), [@lixd](https://github.com/lixd))

### Enhancements & Updates
- Use Helm action package instead of using Helm binary.  ([kubesphere#4852](https://github.com/kubesphere/kubesphere/pull/4852), [@nioshield](https://github.com/nioshield))
- Adjust the priority of bash and sh in the kubectl terminal.([kubesphere#5075](https://github.com/kubesphere/kubesphere/pull/5075), [@tal66](https://github.com/tal66))
- Fix the issue that ks-apiserver cannot start due to DiscoveryAPI exception. ([kubesphere#5408](https://github.com/kubesphere/kubesphere/pull/5408), [@wansir](https://github.com/wansir))
- Fix the issue that the pod status is inconsistent with the filtered status when filtering by status on the pod list page. ([kubesphere#5483](https://github.com/kubesphere/kubesphere/pull/5483), [@frezes](https://github.com/frezes))
- Support querying the secret list according to the secret type by supporting fieldSelector filtering. ([kubesphere#5300](https://github.com/kubesphere/kubesphere/pull/5300), [@nuclearwu](https://github.com/nuclearwu))