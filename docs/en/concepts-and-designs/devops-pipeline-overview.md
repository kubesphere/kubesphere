# KubeSphere DevOps Pipeline Overview

KubeSphere DevOps Pipeline is aimed to meet the complex CI/CD requirements on Kubernetes.

With the KubeSphere DevOps Pipeline, you can quickly construct a complete CI/CD workflow and architecture by manipulating with the KubeSphere console.


## DevOps Pipeline Capabilities

* Build, Deploy, Launch services in Kubernetes 
* Use Kubernetes dynamic agent to release the ability of Kubernetes to dynamically expand
* In SCM Pipeline and Out of SCM Pipeline 
* Easy-to-use graphical pipeline editing panel 
* Multi-tenant isolation


### DevOps Pipeline API

The KubeSphere DevOps Pipeline API will encapsulate the following APIs to provide a standardized REST API:

- Jenkins Core API
- Jenkins BlueOcean API
- Sonarqube API
- Other Plugins API

KubeSphere apiserver will provide multi-tenant API, pipeline API, credential API, code quality analysis API, etc.

![ks-devops-api](../../images/devops-api.png)


### Multi-tenant isolation

In the current version (v2.1.0), multi-tenancy in the DevOps part is done with the ability of the [role-strategy-plugin](https://github.com/jenkinsci/role-strategy-plugin) plugin. KubeSphere will automatically synchronize permission rules in this plugin.

In the future, KubeSphere DevOps will authentication based on [OPA](https://www.openpolicyagent.org/).

### Integration with Jenkins

KubeSphere integrates with standard Jenkins, customizing plugins and configurations.

#### Distribution of plugins

To meet the needs of users in private cloud environments, KubeSphere uses the built-in nginx as a jenkins update center. The jenkins update center is provided as a Docker image + [Helm Chart](https://github.com/kubesphere/ks-installer/tree/master/roles/ks-devops/jenkins-update-center).

#### Jenkins configuration

KubeSphere uses Docker Image, Jenkins update Center and Helm Chart to distribute Jenkins。

The list of plugins and configuration required by Jenkins will be provided by [Helm Chart](https://github.com/kubesphere/ks-installer/tree/master/roles/ks-devops/jenkins).

We use [Groovy Script](https://wiki.jenkins.io/display/JENKINS/Groovy+Hook+Script) and [JCasC](https://github.com/jenkinsci/configuration-as-code-plugin) to initialize Jenkins.


### Pipeline Builder Image And Jenkins PodTemplate

Jenkins does not include any agent configuration by default, KubeSphere provides some default agents (including docker image and podTemplate configuration).

The default agent image will be built based on the [builder base](https://github.com/kubesphere/builder-base), you can search the keyword of repository `builder xxx`  in KubeSphere [github](https://github.com/kubesphere/).

### In SCM Pipeline and Out of SCM Pipeline 

KubeSphere's pipeline syntax will be fully compatible with Jenkins' pipeline syntax. Jenkinsfile found in SCM will be supported with Jenkins plugin.

We will provide a plug-in SCM API, allowing users to graphically edit Jenkinsfile, Dockerfile and other configurations in SCM on KubeSphere.

### Sonarqube Integration

KubeSphere will retrieve Jenkins pipelines that have performed Sonarqube code analysis. And provide API to access analysis report.


### More

If you have more questions, you can create an issue on [github](https://github.com/kubesphere/kubesphere)
