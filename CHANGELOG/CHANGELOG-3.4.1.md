- [v3.4.1](#v341)
  - [App Store](#app-store)
    - [Bug Fixes](#bug-fixes)
  - [Console](#console)
    - [Bug Fixes](#bug-fixes-1)
  - [Multi Cluster](#multi-cluster)
    - [Bug Fixes](#bug-fixes-2)
  - [Observability](#observability)
    - [Bug Fixes](#bug-fixes-3)
  - [Authentication \& Authorization](#authentication--authorization)
    - [Bug Fixes](#bug-fixes-4)
  - [DevOps](#devops)
    - [Bug Fixes](#bug-fixes-5)
    - [Optimization](#optimization)
  - [User Experience](#user-experience)
    - [Bug Fixes](#bug-fixes-6)
  - [Monitoring](#monitoring)
    - [Bug Fixes](#bug-fixes-7)


# v3.4.1

## App Store

### Bug Fixes

- Fix the error on the application repository page. ([console#4188](https://github.com/kubesphere/console/pull/4188), [@ic0xgkk](https://github.com/ic0xgkk), [console#4224](https://github.com/kubesphere/console/pull/4224), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the error in the application approval process. ([kubesphere#5962](https://github.com/kubesphere/kubesphere/pull/5962), [@inksnw](https://github.com/inksnw))
- Support "select" in Appdeploy schemaform.([console#3415](https://github.com/kubesphere/console/pull/3415), [@SongJXin](https://github.com/SongJXin))


## Console

### Bug Fixes

- Fix the issue of losing status when modifying a CRD. ([console#3866](https://github.com/kubesphere/console/pull/3866), [@yzxiu](https://github.com/yzxiu))
- Fix the incorrect style for IsolateInfo and RuleInfo in the Network Policy panel. ([console#4180](https://github.com/kubesphere/console/pull/4180), [@nekomeowww](https://github.com/nekomeowww))
- Fix inaccurate Chinese translations on some pages. ([console#4186](https://github.com/kubesphere/console/pull/4186), [@Hanmo123](https://github.com/Hanmo123), [console#4201](https://github.com/kubesphere/console/pull/4201), [@studyingwang23](https://github.com/studyingwang23), [console#4195](https://github.com/kubesphere/console/pull/4195), [@harrisonliu5](https://github.com/harrisonliu5), [console#4226](https://github.com/kubesphere/console/pull/4226), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the resource creation failure issue in higher versions of Kubernetes. ([console#4190](https://github.com/kubesphere/console/pull/4190), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the issue in the dashboard where the names in Recently Access are obscured. ([console#4210](https://github.com/kubesphere/console/pull/4210), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the issue on the list page where the Customize Columns button displays incomplete content. ([console#4211](https://github.com/kubesphere/console/pull/4211), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the display error of pod status. ([console#4214](https://github.com/kubesphere/console/pull/4214), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the issue on the PVC page where the capacity selection numbers are displayed incorrectly. ([console#4227](https://github.com/kubesphere/console/pull/4227), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the issue when adding containers on the Create Deployment page, using tags to search for images is not working. ([console#4228](https://github.com/kubesphere/console/pull/4228), [@harrisonliu5](https://github.com/harrisonliu5))
- Support building ARM64 images. ([console#4102](https://github.com/kubesphere/console/pull/4102), [@gunine](https://github.com/gunine))

## Multi Cluster

### Bug Fixes

- Fix the issue where CD-related clusters are displayed incorrectly in a multi-cluster environment. ([console#4165](https://github.com/kubesphere/console/pull/4165), [@ks-ci-bot](https://github.com/ks-ci-bot))
- Fix the issue of mistakenly adding a host cluster as a member cluster of another cluster. ([kubesphere#5961](https://github.com/kubesphere/kubesphere/pull/5961), [@ks-ci-bot](https://github.com/ks-ci-bot))


## Observability

### Bug Fixes

- Fix the issue that CPU and memory statistics charts are not displaying. ([console#4182](https://github.com/kubesphere/console/pull/4182), [@fuchunlan](https://github.com/fuchunlan))
- Fix API call errors on the notification channel page. ([console#4183](https://github.com/kubesphere/console/pull/4183), [@fuchunlan](https://github.com/fuchunlan))
- Fix the blank log receiver page. ([console#4184](https://github.com/kubesphere/console/pull/4184), [@fuchunlan](https://github.com/fuchunlan))
- Fix the issue on the new notification channel page where conditional filtering values are missing. ([console#4225](https://github.com/kubesphere/console/pull/4225), [@harrisonliu5](https://github.com/harrisonliu5))


## Authentication & Authorization

### Bug Fixes

- Fix LDAP login failure. ([console#4187](https://github.com/kubesphere/console/pull/4187), [@fuchunlan](https://github.com/fuchunlan))


## DevOps

### Bug Fixes

- Fix the issue where shell is not effective in graphical editing. ([console#4206](https://github.com/kubesphere/console/pull/4206), [@yazhouio](https://github.com/yazhouio))
- When a cluster is not ready or does not install DevOps, DevOps projects are unavailable. ([console#4216](https://github.com/kubesphere/console/pull/4216), [@yazhouio](https://github.com/yazhouio))
- Fix the incorrect parameter passing in Jenkins. ([console#4217](https://github.com/kubesphere/console/pull/4217), [@yazhouio](https://github.com/yazhouio))
- Fix the issue that clicking the replay button pops up an error prompt. ([console#4219](https://github.com/kubesphere/console/pull/4219), [@yazhouio](https://github.com/yazhouio))
- Fix the issue that the details of a pipeline cannot be viewed. ([console#4220](https://github.com/kubesphere/console/pull/4220), [@yazhouio](https://github.com/yazhouio))
-  Fix the run error due to the large DevOps pipeline logs. ([console#4221](https://github.com/kubesphere/console/pull/4221), [@yazhouio](https://github.com/yazhouio))
- Fix Jenkins image vulnerability. ([console#4222](https://github.com/kubesphere/console/pull/4222), [@yazhouio](https://github.com/yazhouio))
- Fix the issue that failed to upgrade DevOps to 3.4.0. ([ks-installer#2247](https://github.com/kubesphere/ks-installer/pull/2247), [@yazhouio](https://github.com/yazhouio))
- Fix the error in the cleanup task. ([ks-DevOps#1014](https://github.com/kubesphere/ks-DevOps/pull/1014), [@yazhouio](https://github.com/yazhouio))
- Fix the failure to set a timeout. ([ks-DevOps#1016](https://github.com/kubesphere/ks-DevOps/pull/1016), [@yazhouio](https://github.com/yazhouio))
- Fix the bug with downloading multi-branch-pipeline artifacts. ([ks-DevOps#973](https://github.com/kubesphere/ks-DevOps/pull/973), [@yazhouio](https://github.com/yazhouio))
- Fix the issue that disabling discarded history pipelineruns doesn't work. ([ks-DevOps#989](https://github.com/kubesphere/ks-DevOps/pull/989), [@yazhouio](https://github.com/yazhouio))
- Fix the issue that some application resources are not deleted when cascade deleting multiple applications. ([ks-DevOps#990](https://github.com/kubesphere/ks-DevOps/pull/990), [@yazhouio](https://github.com/yazhouio))


### Optimization

- Display the git repo link on the Pipeline page. ([console#4215](https://github.com/kubesphere/console/pull/4215), [@yazhouio](https://github.com/yazhouio))
- Improve the API documentation for DevOps. ([ks-DevOps#968](https://github.com/kubesphere/ks-DevOps/pull/968), [@yazhouio](https://github.com/yazhouio))


## User Experience

### Bug Fixes

- Fix the issue on the Statefuls page where the Pod Grace Period parameter is missing. ([console#4207](https://github.com/kubesphere/console/pull/4207), [@yazhouio](https://github.com/yazhouio))
- Fix the issue where the cluster gateway is not displayed in cluster management. ([console#4209](https://github.com/kubesphere/console/pull/4209), [@harrisonliu5](https://github.com/harrisonliu5))
- Fix the error when creating an application route. ([console#4213](https://github.com/kubesphere/console/pull/4213), [@harrisonliu5](https://github.com/harrisonliu5))
- Add pagination for listing repository tags. ([kubesphere#5958](https://github.com/kubesphere/kubesphere/pull/5958), [@ks-ci-bot](https://github.com/ks-ci-bot))


## Monitoring

### Bug Fixes

- Fix the issue that the Monitoring Target field is displayed blank. ([kubesphere#5834](https://github.com/kubesphere/kubesphere/pull/5834), [@junotx](https://github.com/junotx))
