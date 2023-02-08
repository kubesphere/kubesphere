- [v3.3.2](#v332)
  - [DevOps](#devops)
    - [Enhancements \& Upgrades](#enhancements--upgrades)
    - [Bug Fixes](#bug-fixes)
  - [App Store](#app-store)
    - [Bug Fixes](#bug-fixes-1)
  - [Observability](#observability)
    - [Bug Fixes](#bug-fixes-2)
  - [Service Mesh](#service-mesh)
    - [Bug Fixes](#bug-fixes-3)
  - [NetWork](#network)
    - [Enhancements \& Upgrades](#enhancements--upgrades-1)
  - [Storage](#storage)
    - [Enhancements \& Upgrades](#enhancements--upgrades-2)
    - [Bug Fixes](#bug-fixes-4)
  - [Authentication \& Authorization](#authentication--authorization)
    - [Enhancements \& Upgrades](#enhancements--upgrades-3)
    - [Bug Fixes](#bug-fixes-5)
  - [Development \& Testing](#development--testing)
    - [Bug Fixes](#bug-fixes-6)
  - [User Experience](#user-experience)



# v3.3.2

## DevOps

### Enhancements & Upgrades

- Add the latest GitHub Actions. ([ks-devops#879](https://github.com/kubesphere/ks-devops/pull/879), [@pixiake](https://github.com/pixiake))
- Save the PipelineRun results to the configmap. ([ks-devops#855](https://github.com/kubesphere/ks-devops/pull/855), [@LinuxSuRen](https://github.com/LinuxSuRen)), ([ks-devops#887](https://github.com/kubesphere/ks-devops/pull/887), [@yudong2015](https://github.com/yudong2015)), ([ks-devops-helm-chart#83](https://github.com/kubesphere-sigs/ks-devops-helm-chart/pull/83), [@chilianyi](https://github.com/chilianyi))
- Modify the Chinese description of the status of ArgoCD applications. ([console#4011](https://github.com/kubesphere/console/pull/4011), [@Bettygogo2021](https://github.com/Bettygogo2021))
- Add more information to continuous deployment parameters.([console#4074](https://github.com/kubesphere/console/pull/4074), [@yazhouio](https://github.com/yazhouio))
- Add a link for PipelineRun in the aborted state.([console#4029](https://github.com/kubesphere/console/pull/4029), [@yazhouio](https://github.com/yazhouio))
- Add an ID column for PipelineRun, and the ID will be displayed when users run kubectl commands. ([ks-devops#896](https://github.com/kubesphere/ks-devops/pull/896), [@yudong2015](https://github.com/yudong2015))
- Remove the queued state from pipelinerun. ([ks-devops#860](https://github.com/kubesphere/ks-devops/pull/860), [@chilianyi](https://github.com/chilianyi))

### Bug Fixes

- Fix an issue where webhook configurations are missing after users change and save pipeline configurations.([ks-devops#888](https://github.com/kubesphere/ks-devops/pull/888), [@yudong2015](https://github.com/yudong2015))
- Fix an issue where downloading DevOps pipeline artifacts fails. ([console#4036](https://github.com/kubesphere/console/pull/4036), [@yazhouio](https://github.com/yazhouio))
- Fix an issue where the image address does not match when a service is created by using a JAR/WAR file. ([console#4085](https://github.com/kubesphere/console/pull/4085), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix an issue where the status of PipelineRun changes from `Cancelled` to `Not running`. ([ks-devops#896](https://github.com/kubesphere/ks-devops/pull/896), [@yudong2015](https://github.com/yudong2015))
- Fix the automatic cleaning behavior of pipelines, and keep it consistent with the cleaning configuration of Jenkins. ([ks#270](https://github.com/kubesphere-sigs/ks/pull/270), [@yudong2015](https://github.com/yudong2015))


## App Store

### Bug Fixes

- Fix an issue where the application icon is not displayed on the uploaded application template.([kubesphere#5467](https://github.com/kubesphere/kubesphere/pull/5467),[@liangzai006](https://github.com/liangzai006))
- Fix an issue where the homepage of an application is not displayed on the application information page.([kubesphere#5425](https://github.com/kubesphere/kubesphere/pull/5425),[@liangzai006](https://github.com/liangzai006))
- Fix an issue where importing built-in applications fails.([openpitrix-jobs#29](https://github.com/kubesphere/openpitrix-jobs/pull/29),[@liangzai006](https://github.com/liangzai006))
- Fix a UUID generation error in an IPv6-only environment.([kubesphere#5419](https://github.com/kubesphere/kubesphere/pull/5419),[@isyes](https://github.com/isyes))


## Observability

### Bug Fixes

- Fix a parsing error in the config file of logsidecar-injector. ([ks-installer#2154](https://github.com/kubesphere/ks-installer/pull/2154), @junotx),([logsidecar-injector#6](https://github.com/kubesphere/logsidecar-injector/pull/6), @junotx)


## Service Mesh

### Bug Fixes

- Fix an issue that application governance of Bookinfo projects without service mesh enabled is not disabled by default. ([kubesphere#4037](https://github.com/kubesphere/console/pull/4037))
- Fix an issue where the delete button is missing on the blue-green deployment details page. ([kubesphere#4031](https://github.com/kubesphere/console/pull/4031))


## NetWork

### Enhancements & Upgrades

- Restrict network isolation of projects within the current workspace. ([kubesphere#4019](https://github.com/kubesphere/console/pull/4019))

## Storage

### Enhancements & Upgrades

- Display the cluster to which system-workspace belongs in multi-cluster environments. ([kubesphere#4077](https://github.com/kubesphere/console/pull/4077))
- Rename route to ingress. ([klubesphere#4018](https://github.com/kubesphere/console/issues/4018))

### Bug Fixes

- Fix a storage class error of PVCs  on the page for editing federated projects. ([kubesphere#4045](https://github.com/kubesphere/console/pull/4045))


## Authentication & Authorization

### Enhancements & Upgrades

- Add dynamic options for cache. ([kubesphere#5325](https://github.com/kubesphere/kubesphere/pull/5325),[@zhou1203](https://github.com/zhou1203))
- Remove the "Alerting Message Management" permission. ([kubesphere#2150](https://github.com/kubesphere/ks-installer/pull/2150))

### Bug Fixes

- Fix an issue where platform roles with platform management permisions cannot manage clusters. ([kubesphere#5334](https://github.com/kubesphere/kubesphere/pull/5334),[@zhou1203](https://github.com/zhou1203))

## Development & Testing

### Bug Fixes

- Fix an issue where some data is in the `Out of sync` state after the live-reload feature is introduced.([kubesphere#5422](https://github.com/kubesphere/kubesphere/pull/5422),[@hongzhouzi](https://github.com/hongzhouzi))
- Fix an issue where the ks-apiserver fails when it is reloaded multiple times.([kubesphere#5457](https://github.com/kubesphere/kubesphere/pull/5457),[@hongzhouzi](https://github.com/hongzhouzi))
- Fix an issue where caching resources fails if some required CRDs are missing.([kubesphere#5408](https://github.com/kubesphere/kubesphere/pull/5408),[@wansir](https://github.com/wansir)),([kubesphere#5466](https://github.com/kubesphere/kubesphere/pull/5466),[@hongzhouzi](https://github.com/hongzhouzi))
- Fix an ks-apiserver panic error.([kubesphere#5428](https://github.com/kubesphere/kubesphere/pull/5428),[@smartcat999](https://github.com/smartcat999))
- Fix an issue where Goroutine leaks occur when the audit event sender times out. ([kubesphere#5342](https://github.com/kubesphere/kubesphere/pull/5342),[@hzhhong](https://github.com/hzhhong))

## User Experience

- Limit the length of cluster names. ([kubesphere#4059](https://github.com/kubesphere/console/pull/4059))
- Fix an issue where pod replicas of a federated service are not automatically refreshed. ([kubesphere#4066](https://github.com/kubesphere/console/pull/4066))
- Fix an issue where related pods are not deleted after users delete a service. ([kubesphere#4021](https://github.com/kubesphere/console/pull/4021))
- Fix an issue where the number of nodes and roles are incorrectly displayed when there is only one node. ([kubesphere#4032](https://github.com/kubesphere/console/pull/4032))