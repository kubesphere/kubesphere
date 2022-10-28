- [v3.3.1](#v331)
  - [DevOps](#devops)
    - [Enhancements & Updates](#enhancements--updates)
    - [Bug Fixes](#bug-fixes)
  - [Network](#network)
    - [Bug Fixes](#bug-fixes-1)
  - [App Store](#app-store)
    - [Bug Fixes](#bug-fixes-2)
  - [Storage](#storage)
    - [Bug Fixes](#bug-fixes-3)
  - [Observability](#observability)
    - [Bug Fixes](#bug-fixes-4)
  - [Authentication & Authorization](#authentication--authorization)
    - [Bug Fixes](#bug-fixes-5)
  - [KubeEdge Integration](#kubeedge-integration)
    - [Bug Fixes](#bug-fixes-6)
  - [User Experience](#user-experience)
  - [API Changes](#api-changes)

# v3.3.1
## DevOps
### Enhancements & Updates
- Add support for editing the kubeconfig binding mode on the pipeline UI.([console#3864](https://github.com/kubesphere/console/pull/3864), [@harrisonliu5](https://github.com/harrisonliu5))

### Bug Fixes
- Fix an issue where Jenkins updates are not synchronized in real time. ([ks-devops#837](https://github.com/kubesphere/ks-devops/pull/837), [@chilianyi](https://github.com/chilianyi))
- Fix the cron expression check failure of Jenkins. ([ks-devops#784](https://github.com/kubesphere/ks-devops/pull/784), [@LinuxSuRen](https://github.com/LinuxSuRen))
- Fix an issue where users fail to check the CI/CD template.([ks-devops-helm-chart#80](https://github.com/kubesphere-sigs/ks-devops-helm-chart/pull/80), [@chilianyi](https://github.com/chilianyi))
- Remove the `Deprecated` tag from the CI/CD template and replace `kubernetesDeploy` with `kubeconfig binding` at the deployment phase.([ks-devops-helm-chart#81](https://github.com/kubesphere-sigs/ks-devops-helm-chart/pull/81), [@chilianyi](https://github.com/chilianyi))
- Fix an issue where pipeline parameters are not updated in time.([console#3864](https://github.com/kubesphere/console/pull/3864), [@harrisonliu5](https://github.com/harrisonliu5))


## Network
### Bug Fixes
- Fix an issue where users fail to create routing rules in IPv6 and IPv4 dual-stack environments.([console#3604](https://github.com/kubesphere/console/pull/3604), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Update the OpenELB check URL. ([console#3550](https://github.com/kubesphere/console/pull/3550), [@harrisonliu5](https://github.com/harrisonliu5))

## App Store
### Bug Fixes
- Fix an issue where the HTTP cookie is empty while sending the traffic update policy request on the traffic monitor page of an application.([console#3836](https://github.com/kubesphere/console/pull/3836), [@mujinhuakai](https://github.com/mujinhuakai))

## Storage
### Bug Fixes
- Set `hostpath` as a required option when users are mounting volumes. ([console#3478](https://github.com/kubesphere/console/pull/3478), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Update storageClass-accessor, so that deleting storage resources no longer requires accessor validation.([kubesphere#5271](https://github.com/kubesphere/kubesphere/pull/5271),[@f10atin9](https://github.com/f10atin9))

## Observability
### Bug Fixes
- Fix inaccurate time unit of the monitoring metrics.([console#3557](https://github.com/kubesphere/console/pull/3557), [@iawia002](https://github.com/iawia002))
- Change the ratio of `ingress_request_duration_95percentage` to 0.95. ([kubesphere#5132](https://github.com/kubesphere/kubesphere/pull/5132), [@iawia002](https://github.com/iawia002))


## Authentication & Authorization
### Bug Fixes
- Add support for filtering workspace roles using the LabelSelector. ([kubesphere#5162](https://github.com/kubesphere/kubesphere/pull/5162), [@zhou1203](https://github.com/zhou1203))
- Add support for customizing or randomly setting an initial amdin password.([ks-install#2067](https://github.com/kubesphere/ks-installer/pull/2067),[@pixiake](https://github.com/pixiake))
- Delete annotations in `role-template-manage-users`, `role-template-view-members` and `role-template-manage-roles`.([ks-install#2062](https://github.com/kubesphere/ks-installer/pull/2062),[@zhou1203](https://github.com/zhou1203))
- Fix an issue where `cluster-admin` cannot view and manage the configmap, secret, and service account.([ks-install#2082](https://github.com/kubesphere/ks-installer/pull/2082),[@zhou1203](https://github.com/zhou1203))
- Delete role `workspace-manager`.([ks-install#2094](https://github.com/kubesphere/ks-installer/pull/2094),[@zhou1203](https://github.com/zhou1203))
- Add role `platform-self-provisioner`. ([ks-install#2095](https://github.com/kubesphere/ks-installer/pull/2095),[@zhou1203](https://github.com/zhou1203))
- Delete role `users-manager`. ([ks-install#2105](https://github.com/kubesphere/ks-installer/pull/2105),[@zhou1203](https://github.com/zhou1203))
- Block `role-template-manage-groups`. ([ks-install#2122](https://github.com/kubesphere/ks-installer/pull/2122),[@zhou1203](https://github.com/zhou1203))

## KubeEdge Integration
### Bug Fixes
- Change the cluster module key from `kubeedge` to `edgeruntime`.([console#3548](https://github.com/kubesphere/console/pull/3548), [@harrisonliu5](https://github.com/harrisonliu5))

## User Experience
- Optimize image building of KubeSphere and the console.([console#3610](https://github.com/kubesphere/console/pull/3610), [@zt1046656665](https://github.com/zt1046656665))
- Fix an issue where the key is not display in LoadBalancer. ([console#3503](https://github.com/kubesphere/console/pull/3503), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix an issue where no prompt is displayed when users do not fill in key and value in the LoadBalancer access mode.([console#3499](https://github.com/kubesphere/console/pull/3499), [@weili520](https://github.com/weili520))
- Fix an issue where the update time of a service is incorrect on the detail page.([console#3803](https://github.com/kubesphere/console/pull/3803), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Fix inaccurate prompt when users are adding an init container.([console#3561](https://github.com/kubesphere/console/pull/3561), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Fix an issue where an error occurs while users enter Chinese characters in secret. ([console#3774](https://github.com/kubesphere/console/pull/3774), [@moweiwei](https://github.com/moweiwei))
- Add a prompt to remind users to select a language or artifact type when users are building images. ([console#3534](https://github.com/kubesphere/console/pull/3534), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix an issue where the total number of pages is incorrectly displayed.([kubesphere#5201](https://github.com/kubesphere/kubesphere/pull/5201), [@yongxingMa](https://github.com/yongxingMa))
- Fix an issue where the update time of an application is incorrect. ([console#3541](https://github.com/kubesphere/console/pull/3541), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Add support for changing the number of items displayed on each page of a table.([console#3486](https://github.com/kubesphere/console/pull/3486), [@weili520](https://github.com/weili520))
- Add support for batch stopping workloads. ([console#3497](https://github.com/kubesphere/console/pull/3497), [@harrisonliu5](https://github.com/harrisonliu5))
- Add the creator annotation to ensure information displayed on the pod detail page is consistent with other details pages. ([console#3820](https://github.com/kubesphere/console/pull/3820), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Fix a 404 issue in Documentation. ([console#3484](https://github.com/kubesphere/console/pull/3433) [@PrajwalBorkar](https://github.com/PrajwalBorkar))
- Add support for displaying the revision record when the workload type is `statefulsets` or `daemonsets`.([console#3819](https://github.com/kubesphere/console/pull/3819), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Add a support page. ([console#3832](https://github.com/kubesphere/console/pull/3832), [@yazhouio](https://github.com/yazhouio))
- Fix an issue where the status of a cluster remains true when the cluster fails to join Federation. ([kubesphere#5137](https://github.com/kubesphere/kubesphere/pull/5137), [@x893675](https://github.com/x893675))
- Fix an issue where traffic allocation fails in the canary release mode. ([console#3542](https://github.com/kubesphere/console/pull/3542), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Add support for duplicate name validation of containers.  ([console#3559](https://github.com/kubesphere/console/pull/3559), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Add support for duplicate name validation of service names. ([console#3696](https://github.com/kubesphere/console/pull/3696), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Fix an issue where configurations do not take effect when users set the pod request to 0.([console#3827](https://github.com/kubesphere/console/pull/3827), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix an issue where canary release goes wrong when multiple operating systems are selected.([console#3479](https://github.com/kubesphere/console/pull/3479), [@zhaohuiweixiao](https://github.com/zhaohuiweixiao))
- Fix an issue where configmap configurations cannot be saved while users are creating a workload.([console#3416](https://github.com/kubesphere/console/pull/3416),[@weili520](https://github.com/weili520))


## API Changes
- Change the patch type of `PatchWorkspaceTemplate` from `MergePatchType` to `JSONPatchType`.([kubesphere#5217](https://github.com/kubesphere/kubesphere/pull/5217), [@zhou1203](https://github.com/zhou1203))
- Fix the "No Cluster Available" issue during log search. ([console#3555](https://github.com/kubesphere/console/pull/3555), [@harrisonliu5](https://github.com/harrisonliu5))