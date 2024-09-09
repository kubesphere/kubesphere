# v4.1.1

## KubeSphere

### Features

* Refactor based on the new microkernel architecture of KubeSphere LuBan.
* Introduce the KubeSphere Marketplace as a built-in feature.
* Support for managing extensions through the Extensions Center.
* UI and API can be extended.
* Support for one-click import of member clusters via kubeconfig.
* Support for KubeSphere Service Accounts.
* Support for dynamic extension of the Resource API.
* Support for adding clusters, workspaces, and projects to quick access.
* Enabled file upload and download via container terminal.
* Adapted to cloud-native gateways (Kubernetes Ingress API) from different vendors.
* Support API rate limiting.
* Support creating persistent volumes on the page.

### Enhancements

* Support for selecting all clusters when creating a workspace.
* Optimization of web kubectl, supporting dynamic recycling of pods and fuzzy search when switching clusters.
* Optimization of node list, changing the default sorting to ascending order.
* Only allow trusted OAuth clients to verify user identity directly using username and password.
* Streamline the Agent components deployed in member clusters.
* Split some configurations in KubeSphere Config as independent configuration items.
* Adjust the search result of container images to sort in reverse chronological order.
* Support for editing user aliases.
* Display scheduling status in the cluster list.
* Support binaryData data display in ConfigMap details.
* Refactor the Workbench page.
* Remove unnecessary extensions.
* Support fast deployment and uninstallation via helm.
* Simplify the deployment of agents in member clusters.
* Support for disabling terminal for control plane nodes.
* Support triggering cluster resource synchronization proactively.
* User experience improvements on the workload pages under clusters.
* User experience improvements on the application list page.
* User experience improvements on the persistent volume claim and storage class list pages.
* Optimize the display of excessively long resource names.
* Support globally enabling fieldValidation.
* Support horizontal movement on the cluster nodes list page.


### Bug Fixes

* Fix the issue of the node terminal displaying "connecting" indefinitely.
* Fix the potential issue of unauthorized access to resources in the workspace.
* Fix the potential issue of unauthorized access to cluster authorization APIs in the workspace.
* Fix the issue of abnormal session logout due to incorrect configuration.
* Fix the issue of exceptions when adding image service information to pull images from a specified repository.
* Fix the issue of missing ownerReferences when editing Secrets.
* Fix the issue of white screen and incorrect page redirection during the initial login.
* Fix the scrolling issue with checkboxes in Windows environment.
* Fix the problem where the cluster management entry couldn't be found when logged in as cluster-admin.
* Fix the issue where closing the kubectl container terminal did not terminate the corresponding pod.
* Fix the issue where the cluster cannot be selected when downloading kubeconfig.
* Fix the issue where resource names were not fully displayed in some lists due to excessive length.
* Fix the missing translations in some pages.
* Fix the blank display issue on container details page.
* Fix the default skip certificate verification issue when selecting HTTPS for image registry addresses when creating secrets.
* Fix inability to edit project roles for service accounts of member clusters.
* Fix inability to edit settings for configmaps without key-value pairs.
* Fix inability to edit or delete key-value pair data for configmaps of member clusters.
* Fix the display issue with pop-up dialog when removing unready clusters on the cluster management page.
* Fix the progress bar display issue when removing clusters on cluster management page.
* Fix disappearance of the selection status of previously selected clusters after searching for clusters in pop-up dialog for "Add tags to clusters".
* Fix pagination issue for pods on workload details page.
* Fix display of HTML comments in changelogs on extension details page.
* Fix incomplete display of floating elements on list pages under certain circumstances.
* Fix abnormal display of error messages in the top right corner.
* Fix display issue with pop-up dialog for creating workspaces.
* Fix inability to search for Harbor (version 2.8.0 and later) images.
* Fix slow loading of console under HTTPS protocol.
* Specify cluster creator as cluster administrator by default.
* Fix exception when deleting labels from nodes.
* Fix the issue where page does not refresh in real time when adding member clusters.
* Add prompt message when uploading files in containers.
* Fix the issue where clusters are not filtered based on user permissions when selecting clusters.
* Fix potential permission escalation and authorization risks in helm application deployment.
* Fix the issue where file upload freezes when creating application templates.
* Fix the issue where applications created in one project are visible in other projects.
* Fix the issue where bitnami source in application repository cannot be synchronized.
* Fix the issue where application template displays no data.
* Fix the issue where unauthorized users encounter a blank page when deploying applications from the app store.
* Fix incorrect display of types of secrets.
* Fix display issue with workspace list.
* Fix incorrect status in persistent volume list.
* Fix failure to create PVC based on snapshots.
* Remove unnecessary prompt for persistent volume expansion.
* Fix incorrect display of type dropdown when creating secrets.
* Fix data filling error when creating secrets with the "Image registry information" type.
* Fix the issue where workload list cannot retrieve all projects.
* Fix abnormal display of prompt information for pods in workloads.
* Fix the issue where versions displayed in CRDs page is not the latest.
* Fix display issue when searching for clusters in cluster list.
* Fix the issue where web kubectl terminal is unusable in EKS environment.

### API Updates

#### API Removal

The following APIs will be removed in v4.1:

**Multi-cluster**

The multi-cluster proxy request API `/API_PREFIX/clusters/{cluster}/API_GROUP/API_VERSION/...` has been removed. Please use the new multi-cluster proxy request path rule `/clusters/{cluster}/API_PREFIX/API_GROUP/API_VERSION/...` instead.

**Access Control**

- The `iam.kubesphere.io/v1alpha2` API version has been removed. Please use the `iam.kubesphere.io/v1beta1` API version instead.

- Significant changes in `iam.kubesphere.io/v1beta1`:
  The API Group for Role, RoleBinding, ClusterRole, and ClusterRoleBinding resources has changed from `rbac.authorization.k8s.io` to `iam.kubesphere.io`.

**Multi-tenancy**

- Partial APIs in `tenant.kubesphere.io/v1alpha1` and `tenant.kubesphere.io/v1alpha2` API versions have been removed. Please use the `tenant.kubesphere.io/v1beta1` API version instead.

- Significant changes in `tenant.kubesphere.io/v1beta1`:
  `spec.networkIsolation` in `Workspace` has been removed.

**kubectl**

- The `/resources.kubesphere.io/v1alpha2/users/{user}/kubectl` interface has been removed. Terminal-related operations no longer need to call this interface.
- The API path for the user web kubectl terminal has been adjust from `/kapis/terminal.kubesphere.io/v1alpha2/namespaces/{namespace}/pods/{pod}/exec` to `/kapis/terminal.kubesphere.io/v1alpha2/users/{user}/kubectl`.

**Gateway**

The `gateway.kubesphere.io/v1alpha1` API version has been removed.

- The API for querying related gateways of the Ingress configuration has been adjust to `/kapis/gateway.kubesphere.io/v1alpha2/namespaces/{namespace}/availableingressclassscopes`.

#### API Deprecations

The following APIs have been marked as deprecated and will be removed in future versions:

- Cluster validation API
- Config configz API
- OAuth token review API
- Operations job rerun API
- Resources v1alpha2 API
- Resources v1alpha3 API
- Tenant v1alpha3 API
- Legacy version API

### Known Issues

* Upgrade from version 3.x to 4.x is not supported at present, but will be supported in subsequent releases.
* The following functions are temporarily unavailable and will be offered by extensions later:
  * Monitoring
  * Alerting
  * Notifications
  * Istio
  * DevOps
  * Project gateway and cluster gateway
  * Storage volume snapshot
  * Network isolation
  * OpenPitrix for app management


* The following features are currently unavailable and will be supported in subsequent versions:
  * Department management in workspaces

### Misc

- Remove all language options except English and Simplified Chinese by default.
- Remove content related to system components.

