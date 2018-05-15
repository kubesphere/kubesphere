# KubeSphere
[![License](http://img.shields.io/badge/license-apache%20v2-blue.svg)](https://github.com/KubeSphere/KubeSphere/blob/master/LICENSE)

----
***KubeSphere*** is a distribution of [Kubernetes](https://kubernetes.io), aimed to provide quick setup, friendly and easily use, and powerful management features for Kubernetes clusters, which could help both personal and enterprise users, reduce their learning curve of Kubernetes, accelerate their transform process from other container platforms to Kubernetes.   

**Features:**
 - Multiple IaaS platform support, including baremetal/KVM/QingCloud, and more will be supported in future release.  
 - Easy setup of Kubernetes standalone(only one master node) and cluster environment(including High Availability support).  
 - Powerful management console to help business users to manage and monitor the Kubernetes environment.  
 - Integrate with [OpenPitrix](https://github.com/openpitrix) to provide full life cycle of application management and be compatible of helm package.  
 - Support popular open source network solutions, including calico and flannel, also could use [qingcloud hostnic solution](https://github.com/yunify/hostnic-cni) if the Kubernetes is deployed on QingCloud platform.  
 - Support popular open source storage solutions, including Glusterfs and Cephfs, also could use [qingcloud storage solution](https://github.com/yunify/qingcloud-volume-provisioner) if the Kubernetes is deployed on QingCloud platform.  
 - CI/CD support.  
 - Service Mesh support.  
 - Multiple image registries support.  
 - Federation support.  
 - Integrate with QingCloud IAM.  
----

## Motivation

The project originates from the requirement and pains we heard from our customers on public and private QingCloud platform, who have strong will to deploy Kubernetes in their IT system but struggle on completed setup process and long learning curve. With help of KubeSphere, their IT operators could setup Kubernetes environment quickly and use an easy management UI interface to mange their applications.

Getting Started
---------------
**TBD**

## Design

## Contributing to the project

All [members](docs/members.md) of the KubeSphere community must abide by [Code of Conduct](code-of-conduct.md). Only by respecting each other can we develop a productive, collaborative community.

You can then check out how to [setup for development](docs/development.md).


