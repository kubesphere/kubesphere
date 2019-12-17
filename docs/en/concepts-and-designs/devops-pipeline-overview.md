# DevOps Pipeline Overview

KubeSphere Pipeline DevOps aims to complex CI / CD requirements on Kubernetes.

With the KubeSphere DevOps Pipeline, you can complete the construction of the CI / CD process by interacting with the KubeSphere console.


## DevOps Pipeline Capabilities

* Multi-tenant isolation
* Build, Deploy, Launch things in kubernetes 
* Use kubernetes dynamic agent to release the ability of kubernetes to dynamically expand
* In SCM Pipeline and Out of SCM Pipeline 
* Easy-to-use pipeline graphical editing


### DevOps Pipeline API

The KubeSphere DevOps Pipeline API will encapsulate the following APIs (Jenkins Core API / Jenkins BlueOcean API / Sonarqube API / Other Plugins API) to provide a standardized REST API.

KubeSphere apiserver will provide multi-tenant API, pipeline API, credential API, code quality analysis API, etc.

![ks-devops-api](../../images/devops-api.png)


### Multi-tenant isolation

In the current version (v2.1.0), multi-tenancy in the DevOps part is done with the ability of the [role-strategy-plugin](https://github.com/jenkinsci/role-strategy-plugin) plugin. KubeSphere will automatically synchronize permission rules in this plugin.

In the future, kubesphere devops will authentication based on [OPA](https://www.openpolicyagent.org/).

### Integration with Jenkins

KubeSphere integrates with standard Jenkins, customizing plugins and configurations.

#### Distribution of plugins

To meet the needs of users in private cloud environments, KubeSphere uses the built-in nginx as a jenkins update center. The jenkins update center is provided as a Docker image + [Helm Chart](https://github.com/kubesphere/ks-installer/tree/master/roles/ks-devops/jenkins-update-center).

#### Jenkins configuration

KubeSphere uses Docker Image + Jenkins update Center + Helm Chart to distribute Jenkins。

The list of plugins and configuration required by Jenkins will be provided by [Helm Chart](https://github.com/kubesphere/ks-installer/tree/master/roles/ks-devops/jenkins).

We use [Groovy Script](https://wiki.jenkins.io/display/JENKINS/Groovy+Hook+Script) and [JCasC](https://github.com/jenkinsci/configuration-as-code-plugin) to initialize Jenkins.


### Pipeline Builder Image And Jenkins PodTemplate

Jenkins does not include any agent configuration by default, KubeSphere will provide some default agents (including docker image and podTemplate configuration).

The default agent image will be built based on the [builder base](https://github.com/kubesphere/builder-base), you can search `builder xxx` repositories in kubesphere github.

### In SCM Pipeline and Out of SCM Pipeline 

KubeSphere's pipeline syntax will be fully compatible with Jenkins' pipeline syntax. Jenkinsfile found in SCM will be supported with Jenkins plugin.

We will provide a plug-in SCM API, allowing users to graphically edit Jenkinsfile, Dockerfile and other configurations in SCM on KubeSphere.

### Sonarqube integration

KubeSphere will retrieve Jenkins pipelines that have performed Sonarqube code analysis. And provide API to access analysis report.


### More

If you have more questions, you can create an issue on [github](https://github.com/kubesphere/kubesphere)
