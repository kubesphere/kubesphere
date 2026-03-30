---
name: kubesphere-core
description: KubeSphere central controller Skill. Routes to specific Skills based on user requests: multi-cluster management (kubesphere-cluster-management), multi-tenant management (kubesphere-multi-tenant-management), extension management (kubesphere-extension-management). Also provides core architecture, API routing, and API utilities.
---

# KubeSphere Core

## Overview

KubeSphere is a distributed operating system for cloud-native application management built on Kubernetes. Version 4.x adopts a microkernel + extension architecture (codename LuBan) where the core provides essential functions and independent functional modules are delivered as extensions.


## Core Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                              KubeSphere Console                                                 │
│                                   (React-based Web UI + Extension Framework)                                    │
└────────────────────────────────────────────────────────────┬────────────────────────────────────────────────────┘
                                                             │
┌────────────────────────────────────────────────────────────▼────────────────────────────────────────────────────┐
│                                            KubeSphere Core (LuBan)                                              │
│  ┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐  │
│  │ Multi-Cluster     │ │ Multi-Tenant      │ │ K8s Resource      │ │ Extension         │ │ Application       │  │
│  │ Management        │ │ Management        │ │ Management        │ │ Management        │ │ Management        │  │
│  └───────────────────┘ └───────────────────┘ └───────────────────┘ └───────────────────┘ └───────────────────┘  │
└────────────────────────────────────────────────────────────┬────────────────────────────────────────────────────┘
                                                             │
┌────────────────────────────────────────────────────────────▼────────────────────────────────────────────────────┐
│                                      Underlying Kubernetes Clusters                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

| Component | Description |
|-----------|-------------|
| **ks-apiserver** | API aggregation layer. Acts as the central API entry point for KubeSphere, aggregating APIs from core services and extensions. Handles authentication, authorization, and multi-cluster API routing. |
| **ks-controller-manager** | Resource controllers. Manages the lifecycle of KubeSphere custom resources (workspaces, projects, users, etc.). Reconciles desired state and handles event-driven operations. |
| **ks-console** | Web UI. React-based console providing visual management interface for all KubeSphere capabilities. Supports extension framework for custom UI components. |


## API

### API Path Prefixes

KubeSphere provides two main API path prefixes with different routing behaviors:

| Path Prefix | Routed To | Use Case |
|-------------|-----------|----------|
| `/apis/` | kube-apiserver on target cluster | Direct Kubernetes CRUD operations |
| `/kapis/` | KubeSphere API server | KubeSphere-specific APIs, workspace-scoped operations |

### Multi-Cluster API Routing

Use the `/clusters/{cluster-name}/` prefix to forward requests to specific member clusters:

```bash
# Access member-1 cluster
/clusters/member-1/kapis/tenant.kubesphere.io/v1beta1/workspaces/demo/namespaces
```

### API Script Usage

Use the `ks_api.py` script to make API calls. The script handles authentication automatically.

**Prerequisites:**
```bash
pip install requests
export KUBESPHERE_HOST="http://<kubesphere-host>"
python scripts/ks_api.py --login --username admin --password <password>
```

**Usage:**
```bash
# Get token info
python scripts/ks_api.py

# List resources (GET)
python scripts/ks_api.py GET /kapis/tenant.kubesphere.io/v1beta1/workspacetemplates

# Query specific cluster
python scripts/ks_api.py GET /clusters/member-1/kapis/tenant.kubesphere.io/v1beta1/namespaces

# Clear cached token
python scripts/ks_api.py --clear-cache
```


## Skill Routing

`kubesphere-core` routes to specific Skills based on user requests.

| Skill | Capabilities | Restrictions |
|-------|--------------|--------------|
| `kubesphere-core` | Architecture, API routing, API utilities | No management operations |
| `kubesphere-cluster-management` | Multi-cluster management | Cluster operations |
| `kubesphere-multi-tenant-management` | Create user/workspace/project, assign roles | **No** delete, **No** custom roles |
| `kubesphere-extension-management` | Extension install/upgrade/uninstall | Extension-related only |


## References

- [KubeSphere](https://kubesphere.io/)
- [KubeSphere Documentation](https://docs.kubesphere.co/)
- [Extension Development Guide](https://dev-guide.kubesphere.io/extension-dev-guide/en/)
- [API Documentation](https://dev-guide.kubesphere.io/extension-dev-guide/en/references/kubesphere-api-concepts/)
