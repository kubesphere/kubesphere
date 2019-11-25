# Overview

KubeSphere is a multi-tenant cluster management platform which is based on Kubernetes. KubeSphere provides easy-to-use UI to reduce your learning cost for employing the container management platform and make it easier to develop, test and maintain your daily work. It's aimed to resolve the problems relating to Kubernetes’ storage, network, security and its usability.

In addition, the platform has integrated and optimized a range of functional modules which are suitable for containers. KubeSphere provides enterprises with a complete solution for development, automation operation and maintenance, microservice governance, multi-tanency management. Besides, our platform is also able to manage workload, container cluster, service and network, application orchestration as well as database mirroring and storage.

<table>
  <tr>
      <td width="50%" align="center"><b>KubeSphere Dashboard</b></td>
      <td width="50%" align="center"><b>Project Resources</b></td>
  </tr>
  <tr>
     <td><img src="https://pek3b.qingstor.com/kubesphere-docs/png/20190925003707.png"/></td>
     <td><img src="https://pek3b.qingstor.com/kubesphere-docs/png/20190925003504.png"/></td>
  </tr>
  <tr>
      <td width="50%" align="center"><b>CI/CD Pipeline</b></td>
      <td width="50%" align="center"><b>Application Template</b></td>
  </tr>
  <tr>
     <td><img src="https://pek3b.qingstor.com/kubesphere-docs/png/20190925000712.png"/></td>
     <td><img src="https://pek3b.qingstor.com/kubesphere-docs/png/20190925231623.png"/></td>
  </tr>
</table>



## Architecture



KubeSphere adopts the separation of front and back ends, also realizes a cloud native design, the back ends' service components can communicate with external systems through the REST API, see [API documentation](https://kubesphere.io/docs/v2.0/api/kubesphere) for more details. All component are included in the architecture diagram below. KubeSphere can run anywhere from on-premise datacenter to any cloud to edge. In addition, it can be deployed on any Kubernetes distribution.

![](https://pek3b.qingstor.com/kubesphere-docs/png/20190810073322.png)
