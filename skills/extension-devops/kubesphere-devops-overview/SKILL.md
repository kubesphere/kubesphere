---
name: kubesphere-devops-overview
description: Use when working with KubeSphere DevOps extension, CI/CD pipelines, Jenkins integration, or pipeline troubleshooting
---

# KubeSphere DevOps Overview

## Overview

KubeSphere DevOps provides CI/CD capabilities through Jenkins integration, supporting both graphical pipeline editing and Jenkinsfile-based pipelines. It enables automated builds, testing, and deployments across multi-cluster environments with **ArgoCD integration** for GitOps continuous deployment.

## When to Use

- Setting up CI/CD pipelines in KubeSphere
- Configuring Jenkins integration
- Managing DevOps projects and pipelines
- Troubleshooting pipeline execution issues
- Integrating with GitHub, GitLab, or SVN repositories
- Configuring SonarQube for code quality

## Core Concepts

### Resource Mapping

KubeSphere DevOps maps resources across three layers:

```
KubeSphere          Kubernetes                    Jenkins
─────────────────────────────────────────────────────────────
Workspace           Workspace CR                  (authorization)
└── DevOpsProject   ├── DevOpsProject CR          └── Folder
    (Namespace)     └── Namespace (with label)
    └── Pipeline    ├── Pipeline CR               └── WorkflowJob
        └── Run     ├── PipelineRun CR            └── Build #N
```

**Key Concept:** A "DevOps Project" in KubeSphere is fundamentally a **Kubernetes namespace** with the `devops.kubesphere.io/managed=true` label. The DevOpsProject CR exists as a wrapper resource, but when querying for accessible DevOps projects, you interact with **namespaces**, not the DevOpsProject CRs directly.

**For tenants:** Use the `/kapis/devops.kubesphere.io/v1alpha3/workspaces/{workspace}/namespaces` endpoint to list accessible DevOps projects (returns namespace resources). The `/apis/devops.kubesphere.io/v1alpha3/devopsprojects` endpoint requires cluster-scoped permissions and returns 403 for tenants.

### DevOps Project Naming Convention

DevOps projects have two forms of names:

| Name Type | Example | Source | Usage |
|-----------|---------|--------|-------|
| **Shortname** | `devopstest` | `.metadata.generateName` in DevOpsProject CR | Display/user-friendly name |
| **Fullname** | `devopstestc2nj7` | `.metadata.name` in DevOpsProject CR and Namespace | Actual resource identifier |

**Important:**
- The **fullname** is the actual Kubernetes namespace name (e.g., `devopstestc2nj7`)
- When creating a DevOps project with shortname `devopstest`, KubeSphere generates a unique fullname by appending random characters
- All API operations use the **fullname** (namespace name), not the shortname
- When a user refers to a project by shortname and multiple projects match, **ask for confirmation** before proceeding

### Workspace Association

DevOpsProjects belong to a Workspace via label:

```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: DevOpsProject
metadata:
  name: my-project
  labels:
    kubesphere.io/workspace: demo   # Associates with Workspace 'demo'
```

To create and associate:

```bash
# 1. Create Workspace
kubectl apply -f - <<EOF
apiVersion: tenant.kubesphere.io/v1beta1
kind: Workspace
metadata:
  name: demo
EOF

# 2. Create DevOpsProject with label
kubectl apply -f - <<EOF
apiVersion: devops.kubesphere.io/v1alpha3
kind: DevOpsProject
metadata:
  name: my-project
  labels:
    kubesphere.io/workspace: demo
EOF
```

### Project Components

```
┌──────────────────────────────────────────────────────────────┐
│                     DevOps Project                            │
│  (Namespace with devops.kubesphere.io/managed=true label)    │
└──────────────────────┬───────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
┌───────▼──────┐ ┌─────▼─────┐ ┌──────▼──────┐
│  Pipelines   │ │Credentials│ │   Webhooks  │
│              │ │           │ │             │
│ - Graphical  │ │ - SSH     │ │ - GitHub    │
│ - Jenkinsfile│ │ - Basic   │ │ - GitLab    │
│ - Multi-branch│ │ - Token  │ │ - Generic   │
└──────────────┘ └───────────┘ └─────────────┘
```

```
┌──────────────────────────────────────────────────────────────┐
│                     DevOps Project                            │
│  (Namespace with devops.kubesphere.io/managed=true label)    │
└──────────────────────┬───────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
┌───────▼──────┐ ┌─────▼─────┐ ┌──────▼──────┐
│  Pipelines   │ │Credentials│ │   Webhooks  │
│              │ │           │ │             │
│ - Graphical  │ │ - SSH     │ │ - GitHub    │
│ - Jenkinsfile│ │ - Basic   │ │ - GitLab    │
│ - Multi-branch│ │ - Token  │ │ - Generic   │
└──────────────┘ └───────────┘ └─────────────┘
```

## Installation

### Using InstallPlan (Recommended for Production)

**Minimal Installation (Default Config) - RECOMMENDED:**

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: devops
  namespace: kubesphere-system
spec:
  extension:
    name: devops
    version: 1.2.4
  enabled: true
  upgradeStrategy: Manual   # Required for production
  # Note: spec.config is omitted to use extension default values
```

**When to use minimal installation:**
- First-time installation to test default configuration
- Standard deployments without special resource requirements
- When you want to use the extension's tested defaults

**Custom Configuration (Only When Needed):**

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: devops
  namespace: kubesphere-system
spec:
  extension:
    name: devops
    version: 1.2.4
  enabled: true
  upgradeStrategy: Manual   # Required for production
  # config: leave empty to use default values
```

**Custom Configuration (Override Defaults):**

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: devops
  namespace: kubesphere-system
spec:
  extension:
    name: devops
    version: 1.2.4
  enabled: true
  upgradeStrategy: Manual   # Required for production
  config: |
    # Overrides values from DevOps chart's values.yaml
    agent:
      jenkins:
        Master:
          NodeSelector: {}
          resources:
            requests:
              cpu: "500m"
              memory: "4Gi"
            limits:
              cpu: "2000m"
              memory: "8Gi"
        Agent:
          Image: "jenkins/inbound-agent"
          Tag: "3309.v27b_9314fd1a_4-1-jdk21"
          Privileged: false
```

**Important:**
- Always use `upgradeStrategy: Manual` for production
- `config` is optional - omit or leave empty to use extension defaults
- Config values override the extension's `values.yaml` settings
- DevOps is a critical infrastructure component - plan upgrades carefully

### Multi-Cluster Installation

To install DevOps agent on member clusters, add `clusterScheduling`:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: devops
  namespace: kubesphere-system
spec:
  extension:
    name: devops
    version: 1.2.4
  enabled: true
  upgradeStrategy: Manual
  config: |
    # Base config for all clusters
    agent:
      jenkins:
        Master:
          resources:
            requests:
              cpu: "500m"
              memory: "4Gi"
  clusterScheduling:
    placement:
      clusters:
        - host      # Install on host cluster
        - member1   # Install on member1
        - member2   # Install on member2
    # Optional: per-cluster overrides
    overrides:
      member1: |
        agent:
          jenkins:
            Master:
              resources:
                limits:
                  memory: "8Gi"   # Larger master for member1
      member2: |
        agent:
          jenkins:
            Agent:
              NodeSelector:
                zone: west
```

**Key Points:**
- `clusterScheduling.placement.clusters`: List clusters where DevOps agent runs
- `clusterScheduling.overrides`: Cluster-specific config overrides
- Without `clusterScheduling`, DevOps only runs on the host cluster
- Overrides merge with base config, override values take precedence

### Using Helm (Alternative)

```bash
helm upgrade --install devops kse-extensions/devops \
  -n kubesphere-devops-system \
  --create-namespace
```

### Post-Installation Verification

Verify the DevOps installation:

```bash
# Check DevOps pods
kubectl get pods -n kubesphere-devops-system

# Check InstallPlan status
kubectl get installplan devops -n kubesphere-system

# For multi-cluster: check agent status on each cluster
kubectl get installplan devops -n kubesphere-system -o jsonpath='{.status.clusterSchedulingStatuses}'
```

## Architecture

| Component | Purpose | Namespace |
|-----------|---------|-----------|
| devops-jenkins | Jenkins master | kubesphere-devops-system |
| devops-apiserver | DevOps API service | kubesphere-devops-system |
| devops-controller | Resource controllers | kubesphere-devops-system |
| devops-argocd-* | ArgoCD (GitOps) | argocd |
| Jenkins Agent | Pipeline executors | Dynamic (per pipeline) |
|-----------|---------|-----------|
| devops-jenkins | Jenkins master | kubesphere-devops-system |
| devops-apiserver | DevOps API service | kubesphere-devops-system |
| devops-controller | Resource controllers | kubesphere-devops-system |
| Jenkins Agent | Pipeline executors | Dynamic (per pipeline) |
### Jenkins Integration

KubeSphere DevOps integrates with Jenkins for CI/CD execution. The secret `devops-jenkins` contains the admin token for direct Jenkins access:

```bash
# Get Jenkins admin token
TOKEN=$(kubectl -n kubesphere-devops-system get secret devops-jenkins -o jsonpath='{.data.jenkins-admin-token}' | base64 -d)

# Access Jenkins API
kubectl run curl-jenkins --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/api/json"
```

**Jenkins NodePort:**
```bash
kubectl get svc devops-jenkins -n kubesphere-devops-system
# Default: 30180
```

**Access Jenkins Console:**
- URL: `http://<node-ip>:30180`
- Username: `admin`
- Password: Get from secret above

### ArgoCD Integration

KubeSphere DevOps includes **ArgoCD v2.11.7** for GitOps continuous deployment:

**ArgoCD Components:**
| Component | Purpose |
|-----------|---------|
| application-controller | Manages Application state |
| applicationset-controller | Manages ApplicationSet resources |
| dex-server | SSO authentication |
| notifications-controller | Event notifications |
| repo-server | Repository operations |
| argocd-server | API/UI server |
| redis | Cache layer |

**Access ArgoCD:**
```bash
# Get ArgoCD server URL
kubectl get svc devops-agent-argocd-server -n argocd

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d
```

**Key Features:**
- Declarative GitOps deployment
- Multi-source application support
- Automated sync policies
- SSO integration via Dex
- Notification webhooks

## Key Resources
## Key Resources

| Resource | API Version | Purpose |
|----------|-------------|---------|
| Pipeline | devops.kubesphere.io/v1alpha3 | CI/CD pipeline definition |
| DevOpsProject | devops.kubesphere.io/v1alpha3 | DevOps project (namespace wrapper) |
| Credential | v1/Secret | Repository and deployment credentials |

## Quick Commands

```bash
# List DevOps projects
kubectl get devopsprojects

# List pipelines in a project
kubectl get pipelines -n <devops-project-namespace>

# Get pipeline runs
kubectl get pipelineruns -n <devops-project-namespace>

# Check Jenkins status
kubectl -n kubesphere-devops-system get pods -l app=devops-jenkins

# View Jenkins logs
kubectl -n kubesphere-devops-system logs -l app=devops-jenkins

# Get Jenkins admin password
kubectl -n kubesphere-devops-system get secret devops-jenkins -o jsonpath='{.data.jenkins-admin-password}' | base64 -d
```

## Pipeline Types

| Type | Description | Use Case |
|------|-------------|----------|
| **Graphical** | Visual pipeline editor | Simple pipelines, no code |
| **Jenkinsfile (SCM)** | Pipeline defined in repository | Version-controlled pipelines |
| **Jenkinsfile (Inline)** | Pipeline defined in KubeSphere | Quick testing |
| **Multi-branch** | Auto-discovers branches | GitFlow, feature branches |

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Pipeline fails with "No agent" | Check Jenkins agent configuration |
| Cannot access Git repository | Verify credentials and webhook setup |
| kubeconfig credentials fail | Use `string` type instead of `kubeconfigContent` (v1.2+) |
| Jenkins out of memory | Increase Jenkins master resources |
| Pipeline hangs | Check agent pod status and resource limits |

## Version Compatibility

| DevOps | Jenkins | Notes |
|--------|---------|-------|
| v1.2.x | 2.504.1 LTS | kubernetes-cd plugin removed |
| v1.1.x | 2.346.3 LTS | Legacy kubeconfigContent supported |

## References

- [DevOps Documentation](https://docs.kubesphere.io/v4.1/11-use-extensions/02-devops/)
- [Pipeline Syntax](https://www.jenkins.io/doc/book/pipeline/syntax/)
- [Extension README](/root/go/src/github.com/kubesphere/kse-extensions/devops/README.md)
