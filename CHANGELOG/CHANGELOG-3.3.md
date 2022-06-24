- [v3.3.0](#v330)
    - [Changelog since v3.2.1](#changelog-since-v321)
    - [Changes by Area](#changes-by-area)
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

## Changelog since v3.2.1

## Changes by Area

### DevOps

#### Features

- Support GitOps-based continuous deployment, taking Argo CD as the backend.([ks-devops#466](https://github.com/kubesphere/ks-devops/pull/466), [@LinuxSuRen](https://github.com/LinuxSuRen), [console#2990]( https://github.com/kubesphere/console/pull/2990),[@harrisonliu5](https://github.com/harrisonliu5))
- Add a webhook to process pipeline events.([ks-devops#442](https://github.com/kubesphere/ks-devops/pull/442), [@JohnNiang](https://github.com/JohnNiang))
- Support the ability to verify the SCM token directly instead of calling the Jenkins API.([ks-devops#439](https://github.com/kubesphere/ks-devops/pull/439), [@LinuxSuRen](https://github.com/LinuxSuRen))
- Improve the pipeline templates which can be loaded from APIs.([ks-devops#453](https://github.com/kubesphere/ks-devops/pull/453), [@JohnNiang](https://github.com/JohnNiang), [console#2963](https://github.com/kubesphere/console/pull/2963), [@harrisonliu5](https://github.com/harrisonliu5))
- Support for allowlist in basic information of a DevOps project.([console#3224](https://github.com/kubesphere/console/pull/3224), [@harrisonliu5](https://github.com/harrisonliu5))
- Add an Event tab in the pipeline run record.([console#3012](https://github.com/kubesphere/console/pull/3012), [@harrisonliu5](https://github.com/harrisonliu5))
- Avoid using the plaintext secret in the DevOps SCM Verify API.([console#2943](https://github.com/kubesphere/console/pull/2943), [@mangoGoForward](https://github.com/mangoGoForward))
- Add a note stating that s2i and b2i currently do not support the containerd runtime. ([console#2734](https://github.com/kubesphere/console/pull/2734), [@weili520](https://github.com/weili520) )
- Display more repositories and organizations on the Import Code Repository page.([console#3222](https://github.com/kubesphere/console/pull/3222), [@harrisonliu5](https://github.com/harrisonliu5))


### Network

#### Enhancements & Updates

- Integrate OpenELB with KubeSphere for exposing the LoadBalancer type of service.([console#2993](https://github.com/kubesphere/console/pull/2993), [@weili520](https://github.com/weili520))

#### Bug Fixes

- Delete the related gateway after a namespace has been deleted.([kubesphere#4626](https://github.com/kubesphere/kubesphere/pull/4626), [@RolandMa1986](https://github.com/RolandMa1986))
- Avoid log connection leak of gateway pods.([kubesphere#4927](https://github.com/kubesphere/kubesphere/pull/4927), [@qingwave](https://github.com/qingwave))

### App Store

#### Bug Fixes

- Fix ks-controller-manager crashes caused by helm controller NPE.([kubesphere#4602](https://github.com/kubesphere/kubesphere/pull/4602), [@chaunceyjiang](https://github.com/chaunceyjiang))


### Multi-tenancy & Multi-cluster

#### Features

- Remind users when the cluster's kubeconfig certificate is about to expire.([console#3038](https://github.com/kubesphere/console/pull/3038), [@harrisonliu5](https://github.com/harrisonliu5), [kubesphere#4584](https://github.com/kubesphere/kubesphere/pull/4584), [@iawia002](https://github.com/iawia002))
- Display the name of the current cluster in the kubesphere-config configmap.([kubesphere#4679](https://github.com/kubesphere/kubesphere/pull/4679), [@iawia002](https://github.com/iawia002))
- Add the update cluster kubeconfig API.([kubesphere#4562](https://github.com/kubesphere/kubesphere/pull/4562), [@iawia002](https://github.com/iawia002))
- Support cluster member management.([console#3061](https://github.com/kubesphere/console/pull/3061), [@harrisonliu5](https://github.com/harrisonliu5))

### Observability

#### Enhancements & Updates

- Add process/thread monitoring metrics for containers.([kubesphere#4711](https://github.com/kubesphere/kubesphere/pull/4711), [@junotx](https://github.com/junotx))
- Add usage/utilization monitoring metrics for node devices.([kubesphere#4705](https://github.com/kubesphere/kubesphere/pull/4705), [@junotx](https://github.com/junotx))
- Adapt metrics query to upgraded monitoring components.([kubesphere#4621](https://github.com/kubesphere/kubesphere/pull/4621), [@junotx](https://github.com/junotx))
- Support the ability to import Grafana templates as namespaced scope dashboards.([kubeshere#4446](https://github.com/kubesphere/kubesphere/pull/4446), [@zhu733756](https://github.com/zhu733756), [console#3202](https://github.com/kubesphere/console/pull/3202) [@weili520](https://github.com/weili520))
- Support separate definition of data retention period for logging, events, and auditing.([ks-installer#1915](https://github.com/kubesphere/ks-installer/pull/1915), [@wenchajun](https://github.com/wenchajun))
- Upgrade Alertmanager from v0.21.0 to v0.23.0.
- Upgrade Grafana from 7.4.3 to 8.3.3.
- Upgrade Kube-state-metrics from v1.9.7 to v2.3.0.
- Upgrade Node-exporter from v0.18.1 to v1.3.1.
- Upgrade Prometheus from v2.26.0 to v2.34.0.
- Upgrade Prometheus-operator from v0.43.2 to v0.55.1.
- Upgrade Kube-rbac-proxy from v0.8.0 to v0.11.0.
- Upgrade Configmap-reload from v0.3.0 to v0.5.0.
- Upgrade Thanos from v0.18.0 to v0.25.2.
- Upgrade Kube-events from v0.3.0 to v0.4.0.
- Upgrade Fluentbit Operator from v0.11.0 to v0.13.0.
- Upgrade Fluent-bit from v1.8.3 to v1.8.11.

### Authentication & Authorization

#### Features

- Support the ability to disable/enable accounts manually.([kubesphere#4695](https://github.com/kubesphere/kubesphere/pull/4695), [@wansir](https://github.com/wansir), [console#3014](https://github.com/kubesphere/console/pull/3014), [@harrisonliu5](https://github.com/harrisonliu5))

### Storage

#### Features

- pvc-autoresizer
  The user can set the pvc autoresizer policy for the storage class as needed, so that when the remaining capacity of the user's pvc is insufficient, it will be expanded according to the preset policy, and it can help the storage class that only supports offline expansion to automatically start and stop development.([kubesphere#4660](https://github.com/kubesphere/kubesphere/pull/4660), [@f10atin9](https://github.com/f10atin9), [console#3056](https://github.com/kubesphere/console/pull/3056), [@weili520](https://github.com/weili520))
- Snapshot-related resource management
  Users can now manage snapshot content and snapshot classes on the console, and some new pages are available to show details.([kubesphere#4596](https://github.com/kubesphere/kubesphere/pull/4596), [@f10atin9](https://github.com/f10atin9),
  [console#3051](https://github.com/kubesphere/console/pull/3051), [@weili520](https://github.com/weili520))
- StorageClass-accessor
  Users can control the creation/deletion of PVC operations in namespaces and workspaces from the storage class details page. Creation actions that do not comply with the rules will be rejected.([kubesphere#4770](https://github.com/kubesphere/kubesphere/pull/4770), [@f10atin9](https://github.com/f10atin9), [console#3069](https://github.com/kubesphere/console/pull/3069), [@weili520](https://github.com/weili520))
- Add disk usage per hard disk.([console#3063](https://github.com/kubesphere/console/pull/3063), [@harrisonliu5](https://github.com/harrisonliu5))
- Update the cluster's volume snapshot page.([console#2958](https://github.com/kubesphere/console/pull/2958), [@weili520](https://github.com/weili520))


### Service Mesh

#### Bug Fixes

- Resolve port conflict in virtual service which uses multi protocols.([kubesphere#4560](https://github.com/kubesphere/kubesphere/pull/4560), [@RolandMa1986](https://github.com/RolandMa1986))


### KubeEdge Integration

#### Features
- Add shell access to cluster nodes, including edge nodes.([kubesphere#4579](https://github.com/kubesphere/kubesphere/pull/4579), [@lynxcat](https://github.com/lynxcat), [console#2888](https://github.com/kubesphere/console/pull/2888), [@lynxcat](https://github.com/lynxcat) )

#### Enhancements & Updates
- Upgrade KubeEdge from v1.7.2 to v1.9.2.
- Remove EdgeWatcher as KubeEdge v1.9.2 provides similar functions.

### Development & Testing

#### Features

- Add the --controllers flag in ks-controller-manager.
  Now we can enable/disable controllers to greatly reduce resource usage when debugging.([kubesphere#4512](https://github.com/kubesphere/kubesphere/pull/4512), [@live77](https://github.com/live77))
- Support live-reload when the config file has been changed.([kubesphere#4659](https://github.com/kubesphere/kubesphere/pull/4659), [@x893675](https://github.com/x893675))
- Add an agent to report additional information in debugging mode.([kubesphere#4928](https://github.com/kubesphere/kubesphere/pull/4928), [@xyz-li](https://github.com/xyz-li))


### User Experience

- Add more support for language configuration.([console#2782](https://github.com/kubesphere/console/pull/2782), [@xuliwenwenwen](https://github.com/xuliwenwenwen))
- Add life management of containers in workloads.([console#2940](https://github.com/kubesphere/console/pull/2940), [@harrisonliu5](https://github.com/harrisonliu5))
- Support the ability to load the entire configmap as environment variables.([console#3044](https://github.com/kubesphere/console/pull/3044), [@weili520](https://github.com/weili520))
- Add a prompt to enable audit logs.([console#3062](https://github.com/kubesphere/console/pull/3062), [@harrisonliu5](https://github.com/harrisonliu5))
- Add a placeholder at the end of the text box.([console#3200](https://github.com/kubesphere/console/pull/3200), [@weili520](https://github.com/weili520))
- Add cluster view permission when users add a cluster.([console#3296](https://github.com/kubesphere/console/pull/3296), [@harrisonliu5](https://github.com/harrisonliu5))
- Improve the Service Topology Details layout.([console#2945](https://github.com/kubesphere/console/pull/2945), [@tracer1023](https://github.com/tracer1023))
- Enable user password pattern checking.(as same as the console does) ([kubesphere#4481](https://github.com/kubesphere/kubesphere/pull/4481), [@live77](https://github.com/live77))
- Fix the issue "Failed to create a statefulSet with a volume template added".([console#2730](https://github.com/kubesphere/console/pull/2730), [@weili520](https://github.com/weili520) )
- Fix the issue "Application deployment fails if we click too fast".([console#2735](https://github.com/kubesphere/console/pull/2735), [@weili520](https://github.com/weili520) )
- Change volume access mode from ROM to ROX.([console#2751](https://github.com/kubesphere/console/pull/2751), [@123liubao](https://github.com/123liubao))
- Fix the issue "Incorrect information for exceeding the resource limit".( [console#2809](https://github.com/kubesphere/console/pull/2809) , [@weili520](https://github.com/weili520))
- Add time select in fedproject applications traffic.([console#3195](https://github.com/kubesphere/console/pull/3195), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the issue "webhook token and auth should be required fields".([console#2903](https://github.com/kubesphere/console/pull/2903), [@xuliwenwenwen](https://github.com/xuliwenwenwen))
- Improve user experience of member clusters on the console.([kubesphere#4721](https://github.com/kubesphere/kubesphere/pull/4721), [@iawia002](https://github.com/iawia002), [console#3031](https://github.com/kubesphere/console/pull/3031), [@weili520](https://github.com/weili520))
- Optimize the UI texts of deleting a cluster, workspace, and project.([console#3004](https://github.com/kubesphere/console/pull/3004), [@weili520](https://github.com/weili520))
- Fix the issue "Service topology map container monitoring's data doesn't change with the data request".([console#3117](https://github.com/kubesphere/console/pull/3117),[@weili520](https://github.com/weili520))
- Fix the issue "The data is empty when modifying the stateful service".([console#2845](https://github.com/kubesphere/console/pull/2845), [@weili520](https://github.com/weili520))
- Fix the issue "Users cannot customize the time interval for the traffic monitoring on the composed app's details page".([console#2916]( https://github.com/kubesphere/console/pull/2916), [@weili520](https://github.com/weili520) )
- Optimize the Service Type and External Access columns of the service list.([console#3058](https://github.com/kubesphere/console/pull/3058),[@weili520](https://github.com/weili520))
- Fix the issue "The probe tooltip is not displayed when it has data".([console#3213](https://github.com/kubesphere/console/pull/3213), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the issue "Wrong name of containers in service details in Service Topology". ([console#3276](https://github.com/kubesphere/console/pull/3276), [@weili520](https://github.com/weili520))
- Fix the worker node statistics error.([console#3279](https://github.com/kubesphere/console/pull/3279), [@weili520](https://github.com/weili520))
- Fix the workspace API error due to incorrect cluster role.([console#3297](https://github.com/kubesphere/console/pull/3297), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the issue "The data is incorret when twice update the cluster kubeconfig".([console#3392](https://github.com/kubesphere/console/pull/3392), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix goroutine leak while opening the web terminal.([kubesphere#4918](https://github.com/kubesphere/kubesphere/pull/4918), [@anhoder](https://github.com/anhoder))
- Fix ks-apiserver unexpected panic caused by the index out of range error while calling the metrics API.([kubesphere#4691](https://github.com/kubesphere/kubesphere/pull/4691), [@larryliuqing](https://github.com/larryliuqing))
- Fix ks-apiserver crashes caused by resource discovery failed.([kubesphere#4835](https://github.com/kubesphere/kubesphere/pull/4835),  [@wansir](https://github.com/wansir))
- Fix the issue "Kubeconfig generated with the wrong type header".([kubesphere#4936](https://github.com/kubesphere/kubesphere/pull/4936), [@xyz-li](https://github.com/xyz-li))
- Fix the registry verification API to skip TLS verification for self-signed certificates.([kubesphere#4678](https://github.com/kubesphere/kubesphere/pull/4678), [@wansir](https://github.com/wansir))

## API Changes

- With `externalKubeAPIEnabled=true` and `connection.type=proxy`, tower will create the service with the `LoadBlancer` type, and content in annotation with key `tower.kubesphere.io/external-lb-service-annotations` will be applied to the service annotations as k-v, so that users can control how `ccm` processes the service.([kubesphere#4528](https://github.com/kubesphere/kubesphere/pull/4528), [@lxm](https://github.com/lxm))
- Provide RESTful APIs for ClusterTemplate.([ks-devops#468](https://github.com/kubesphere/ks-devops/pull/468), [@JohnNiang](https://github.com/JohnNiang))
- Provide Template related APIs.([ks-devops#460](https://github.com/kubesphere/ks-devops/pull/460), [@JohnNiang](https://github.com/JohnNiang))
- Change KubeEdge proxy service to `http://edgeservice.kubeedge.svc/api/`.([kubesphere#4478](https://github.com/kubesphere/kubesphere/pull/4478), [@zhu733756](https://github.com/zhu733756))




