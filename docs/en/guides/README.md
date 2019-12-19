# Guides
There are many ways that you can help the KubeSphere community.

- Go through our documents, point out or fix unclear things. Translate the documents to other languages.
- Install our [releases](https://kubesphere.io/en/install), try to manage your `kubernetes` cluster with `kubesphere`, and feedback to us about 
what you think.
- Read our source codes, Ask questions for details.
- Find kubesphere bugs, [submit issue](https://github.com/kubesphere/kubesphere/issues), and try to fix it.
- Find kubesphere installer bugs, [submit issue](https://github.com/kubesphere/ks-installer/issues), and try to fix it.
- Find [help wanted issues](https://github.com/kubesphere/kubesphere/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22),
which are good for you to start.
- Submit issue or start discussion through [GitHub issue](https://github.com/kubesphere/kubesphere/issues/new).
- See all forum discussion through [website](https://kubesphere.io/forum/).

## Contact Us
All the following channels are open to the community, you could choose the way you like.
* Submit an [issue](https://github.com/kubesphere/kubesphere/issues)
* See all forum discussion through [website](https://kubesphere.io/forum/).
* Join us at [Slack Channel](https://join.slack.com/t/kubesphere/shared_invite/enQtNTE3MDIxNzUxNzQ0LTZkNTdkYWNiYTVkMTM5ZThhODY1MjAyZmVlYWEwZmQ3ODQ1NmM1MGVkNWEzZTRhNzk0MzM5MmY4NDc3ZWVhMjE).

## For code developer

1. Learn about kubesphere's Concepts And Designs and how to build kubesphere locally  
For developers, first step, read [Concepts And Designs](../concepts-and-designs/README.md) and [Compiling Guide](How-to-build.md). 
Concepts And Designs describes the role of each component in kubesphere and the relationship between them.
Compiling Guide teaches developer how to build the project in local and set up the environment.

2. Understand the workflow of kubesphere development.  
Read [Development Workflow](Development-workflow.md).

3. Understand the best practices for submitting PR and our code of conduct  
Read [best practices for submitting PR](pull-requests.md).  
Read [code of conduct](code-of-conduct.md).

### KubeSphere Core developer

### KubeSphere Installer developer

### UI developer

TODO: UI opensource is on the way

### KubeSphere Application Management developer

TODO(@pengcong)

### KubeSphere Service Mesh developer

TODO(@zryfish)

### Porter developer

TODO(@magicsong)

### KubeSphere Networking developer

### KubeSphere DevOps developer

TODO(@runzexia)

### KubeSphere S2I/B2I developer

TODO(@soulseen)

### KubeSphere Monitoring developer

1. Read kubesphere's [Concepts And Designs for Monitoring](../concepts-and-designs/kubesphere-monitoring.md). Understand KubeSphere's monitoring stack.
2. For Prometheus and its wider eco-system setup, go to [kube-prometheus](https://github.com/kubesphere/prometheus-operator/tree/ks-v0.27.0/contrib/kube-prometheus).
3. For KubeSphere builtin metric rules, see [metrics_rules.go](https://github.com/kubesphere/kubesphere/blob/master/pkg/models/metrics/metrics_rules.go) and [kubernetes-mixin](https://github.com/kubesphere/kubernetes-mixin/blob/ks-v0.27.0/rules/rules.libsonnet).
4. For developers who are interested in KubeSphere monitoring backend, read [Development Guide for Monitoring](kubesphere-monitoring-development-guide.md) and [API doc](https://kubesphere.com.cn/docs/v2.1/api/kubesphere#tag/Cluster-Metrics).

### KubeSphere Logging developer

TODO(@huanggze)

### KubeSphere Altering developer

TODO

### KubeSphere Notification developer

TODO



