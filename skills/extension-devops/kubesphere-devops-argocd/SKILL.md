---
name: kubesphere-devops-argocd
description: Use when configuring ArgoCD in KubeSphere DevOps, including GitOps deployments, application management, SSO setup, or troubleshooting ArgoCD issues
---

# KubeSphere DevOps ArgoCD Configuration

## Overview

KubeSphere DevOps includes **ArgoCD v2.11.7** as a bundled subchart for GitOps continuous deployment. ArgoCD follows the declarative GitOps pattern, automatically syncing application state with Git repositories.

## When to Use

- Setting up GitOps continuous deployment
- Configuring ArgoCD applications and ApplicationSets
- Enabling SSO authentication via Dex
- Managing multi-cluster deployments
- Troubleshooting ArgoCD sync issues
- Configuring repository credentials

## KubeSphere GitOps Integration

### Two Ways to Create Applications

**1. Direct ArgoCD Application (Admin Only)**
- Created in `argocd` namespace
- Requires access to ArgoCD namespace
- Full control over ArgoCD configuration

**2. KubeSphere GitOps Application (Tenant-Friendly)**
- Created via `/kapis/gitops.kubesphere.io/v1alpha1/`
- Application CR created in tenant namespace
- KubeSphere automatically creates corresponding ArgoCD Application
- Tenant doesn't need direct ArgoCD access

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Tenant Namespace                          │
│  ┌──────────────────────────────────────────────┐           │
│  │  Application (gitops.kubesphere.io/v1alpha1) │           │
│  │  - Created by tenant via KubeSphere API      │           │
│  │  - Label: gitops.kubesphere.io/argocd-location: argocd │  │
│  └──────────────────┬───────────────────────────┘           │
└──────────────────────┼──────────────────────────────────────┘
                       │
                       │ KubeSphere Controller watches
                       │ and creates corresponding
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    argocd Namespace                          │
│  ┌──────────────────────────────────────────────┐           │
│  │  Application (argoproj.io/v1alpha1)          │           │
│  │  - Created automatically by KubeSphere       │           │
│  │  - References tenant namespace as target     │           │
│  └──────────────────┬───────────────────────────┘           │
└──────────────────────┼──────────────────────────────────────┘
                       │
                       │ ArgoCD Controller reconciles
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    Tenant Namespace                          │
│  ┌──────────────────────────────────────────────┐           │
│  │  Deployed Resources (Pods, Services, etc.)   │           │
│  │  - Created by ArgoCD                         │           │
│  │  - Managed via GitOps                        │           │
│  └──────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

### Key Differences

| Aspect | Direct ArgoCD | KubeSphere GitOps |
|--------|---------------|-------------------|
| **Namespace** | `argocd` | Tenant's namespace |
| **Access Required** | ArgoCD namespace | Tenant namespace only |
| **API Endpoint** | N/A (kubectl) | `/kapis/gitops.kubesphere.io/v1alpha1/` |
| **Tenant Can Create** | ❌ No | ✅ Yes |
| **ArgoCD UI Access** | ✅ Yes | ❌ No (transparent) |
| **Multi-tenancy** | Shared | Isolated per tenant |

## Architecture

### ArgoCD Components

| Component | Pod Name Pattern | Purpose |
|-----------|------------------|---------|
| Application Controller | `devops-agent-argocd-application-controller-*` | Reconciles Application state |
| ApplicationSet Controller | `devops-agent-argocd-applicationset-controller-*` | Manages ApplicationSet CRDs |
| Dex Server | `devops-agent-argocd-dex-server-*` | SSO authentication proxy |
| Notifications Controller | `devops-agent-argocd-notifications-controller-*` | Event notifications |
| Redis | `devops-agent-argocd-redis-*` | Cache and state storage |
| Repo Server | `devops-agent-argocd-repo-server-*` | Git repository operations |
| ArgoCD Server | `devops-agent-argocd-server-*` | API and UI |

**Namespace:** `argocd` (configurable via `argocd.namespace`)

## Installation & Verification

### Check ArgoCD Status

```bash
# Verify ArgoCD namespace exists
kubectl get ns argocd

# Check all ArgoCD pods
kubectl get pods -n argocd

# Check ArgoCD services
kubectl get svc -n argocd
```

### Access ArgoCD UI

```bash
# Get ArgoCD server service
kubectl get svc devops-agent-argocd-server -n argocd

# Port-forward for local access
kubectl port-forward svc/devops-agent-argocd-server -n argocd 8080:443

# Get initial admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d

# Access via: https://localhost:8080
# Username: admin
# Password: (from command above)
```

## Configuration

### Enable/Disable ArgoCD

In DevOps InstallPlan:

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
    agent:
      argocd:
        enabled: true           # Enable ArgoCD
        namespace: "argocd"     # ArgoCD namespace
```

### Custom ArgoCD Configuration

```yaml
config: |
  agent:
    argocd:
      enabled: true
      namespace: "argocd"
      # Full ArgoCD Helm values available
      # See: kse-extensions/devops/charts/agent/charts/argo-cd/values.yaml
      configs:
        cm:
          url: "https://argocd.example.com"
          admin.enabled: "true"
```

## Managing Applications

### Create an Application

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/example/repo.git
    targetRevision: HEAD
    path: k8s/overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: my-app-namespace
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```


### Create via KubeSphere GitOps API (Tenant Method)

**For tenants** who don't have access to the `argocd` namespace:

```bash
# Authenticate as tenant
export API_TOKEN="<tenant-kubesphere-token>"
export KUBESPHERE_API="https://kubesphere-api.example.com"
export DEVOPS_PROJECT="demo-project"

# Create GitOps Application via API
curl -s -X POST "${KUBESPHERE_API}/kapis/gitops.kubesphere.io/v1alpha1/namespaces/${DEVOPS_PROJECT}/applications" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "gitops.kubesphere.io/v1alpha1",
    "kind": "Application",
    "metadata": {
      "name": "guestbook",
      "namespace": "'${DEVOPS_PROJECT}'",
      "labels": {
        "gitops.kubesphere.io/argocd-location": "argocd"
      }
    },
    "spec": {
      "argoApp": {
        "spec": {
          "project": "default",
          "source": {
            "repoURL": "https://github.com/stoneshi-yunify/argocd-example-apps",
            "targetRevision": "HEAD",
            "path": "guestbook"
          },
          "destination": {
            "server": "https://kubernetes.default.svc",
            "namespace": "'${DEVOPS_PROJECT}'"
          },
          "syncPolicy": {
            "automated": {
              "prune": true,
              "selfHeal": true
            },
            "syncOptions": [
              "CreateNamespace=true"
            ]
          }
        }
      }
    }
  }'
```

**What happens:**
1. Tenant creates `Application` (gitops.kubesphere.io/v1alpha1) in their namespace
2. KubeSphere automatically adds label: `gitops.kubesphere.io/argocd-location: argocd`
3. KubeSphere controller creates corresponding ArgoCD Application in `argocd` namespace
4. ArgoCD syncs the application to the tenant's namespace

**Verify Application Status:**

```bash
# Method 1: Check status labels (tenant accessible)
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/gitops.kubesphere.io/v1alpha1/namespaces/demo-project/applications/guestbook" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '{
  health: .metadata.labels["gitops.kubesphere.io/health-status"],
  sync: .metadata.labels["gitops.kubesphere.io/sync-status"]
}'
# Expected output when healthy and synced:
# {
#   "health": "Healthy",
#   "sync": "Synced"
# }

# Method 2: Check detailed status via .status.argoApp (JSON string)
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/gitops.kubesphere.io/v1alpha1/namespaces/demo-project/applications/guestbook" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.status.argoApp' | jq -r '{
  syncStatus: .sync.status,
  healthStatus: .health.status,
  resources: [.resources[] | {kind: .kind, name: .name, status: .status}]
}'

# Check ArgoCD Application (admin only)
kubectl get application -n argocd | grep guestbook
```

**Note:** When `spec.argoApp.spec.destination.server` is `https://kubernetes.default.svc` and `destination.name` is empty or `in-cluster`, the Application deploys to the cluster specified in the API path (e.g., `/clusters/member-1` → member-1 cluster). Tenants should verify deployment via Application status labels as they may not have permissions to query the destination namespace directly.

**Note:** This method requires KubeSphere GitOps controller to be running.


**⚠️ CRITICAL: Required Label**

The Application **MUST** have the label `gitops.kubesphere.io/argocd-location: argocd`. Without this label:
- The controller will **silently ignore** the Application
- No ArgoCD Application will be created
- The Application status will remain **Unknown**

**Evidence from controller logs:**
```
Warning  Invalid  application/private-guestbook  
Cannot find the namespace of the Argo CD instance from key: gitops.kubesphere.io/argocd-location
```

**⚠️ WARNING: Don't Create ArgoCD Application Manually**

When using KubeSphere GitOps Application, the controller automatically creates the corresponding ArgoCD Application. **Do NOT** create an ArgoCD Application manually with the same name or targeting the same resources - this will cause a resource conflict.

**Resource Conflict Example:**
```
Deployment/guestbook-ui is part of applications 
argocd/private-guestbook and stone-devops-private-guestbook
```

This results in:
- SharedResourceWarning
- OutOfSync status
- Conflicting management

### Create an ApplicationSet

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
    - git:
        repoURL: https://github.com/example/repo.git
        revision: HEAD
        directories:
          - path: apps/*
  template:
    metadata:
      name: '{{path.basename}}'
    spec:
      project: default
      source:
        repoURL: https://github.com/example/repo.git
        targetRevision: HEAD
        path: '{{path}}'
      destination:
        server: https://kubernetes.default.svc
        namespace: '{{path.basename}}'
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
```

## Repository Management

### Add a Git Repository

**Via CLI:**
```bash
# Login to argocd CLI
argocd login localhost:8080 --username admin --password $(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)

# Add repository
argocd repo add https://github.com/example/repo.git \
  --username <user> \
  --password <token>
```

**Via Secret:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: repo-github-example
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  type: git
  url: https://github.com/example/repo.git
  username: <username>
  password: <personal-access-token>
```

### Add a Helm Repository

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: repo-helm-stable
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  type: helm
  url: https://charts.helm.sh/stable
  name: stable
```

## SSO Configuration

### Enable Dex for SSO

ArgoCD includes Dex for SSO integration:

```yaml
config: |
  agent:
    argocd:
      configs:
        cm:
          url: https://argocd.example.com
          dex.config: |
            connectors:
              - type: github
                id: github
                name: GitHub
                config:
                  clientID: $dex.github.clientId
                  clientSecret: $dex.github.clientSecret
                  orgs:
                    - name: your-org
```

### Configure Secrets for Dex

```bash
# Create secret for Dex connector credentials
kubectl -n argocd create secret generic argocd-dex-github \
  --from-literal=dex.github.clientId=<client-id> \
  --from-literal=dex.github.clientSecret=<client-secret>
```

## CLI Operations

### Install ArgoCD CLI

```bash
# Download CLI
curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64

# Install
sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd
rm argocd-linux-amd64
```

### Common CLI Commands

```bash
# Login
argocd login <argocd-server-host>

# List applications
argocd app list

# Get application status
argocd app get <app-name>

# Sync application
argocd app sync <app-name>

# Sync with pruning
argocd app sync <app-name> --prune

# Watch sync progress
argocd app wait <app-name> --health

# Rollback
argocd app rollback <app-name> <revision>
```

## Practical Examples

### Complete Workflow: Deploy Guestbook Application

**1. Create Namespace (if needed):**
```bash
kubectl create ns argo-guestbook
```

**2. Create Application:**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: guestbook
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/stoneshi-yunify/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: argo-guestbook
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

**Apply:**
```bash
kubectl apply -f guestbook-app.yaml
```

**3. Check Application Status:**
```bash
# Quick status
kubectl get applications.argoproj.io guestbook -n argocd

# Detailed status
kubectl get applications.argoproj.io guestbook -n argocd -o custom-columns=\
SYNC:.status.sync.status,HEALTH:.status.health.status,REVISION:.status.sync.revision

# Resource health
kubectl get applications.argoproj.io guestbook -n argocd -o jsonpath='{.status.resources}'

# Check deployed resources
kubectl get all -n argo-guestbook
```

**Expected Output:**
```
NAME        SYNC STATUS   HEALTH STATUS
guestbook   Synced        Healthy

SYNC     HEALTH    REVISION
Synced   Healthy   335cffbb730e59b165c308b98c3fa4037822bf2b
```

### Force Resync Application

When you need to force ArgoCD to re-sync (e.g., after manually deleting resources):

**Method 1: Add Refresh Annotation**
```bash
kubectl patch applications.argoproj.io guestbook -n argocd --type merge \
  -p '{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"hard"}}}'
```

**Method 2: Update Spec (triggers reconciliation)**
```bash
kubectl patch applications.argoproj.io guestbook -n argocd --type merge \
  -p '{"spec":{"revisionHistoryLimit":10}}'
```

**With Automated Sync Enabled:**
If `syncPolicy.automated.selfHeal: true`, ArgoCD will automatically recreate deleted resources.

**Test Force Resync:**
```bash
# 1. Delete all resources manually
kubectl delete all --all -n argo-guestbook

# 2. Trigger resync
kubectl patch applications.argoproj.io guestbook -n argocd --type merge \
  -p '{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"hard"}}}'

# 3. Verify resources recreated
kubectl get all -n argo-guestbook
```

### Delete Application

**Delete Application (resources remain by default):**
```bash
kubectl delete applications.argoproj.io guestbook -n argocd
```

**Delete Application and Cleanup Resources:**
```bash
# Delete application
kubectl delete applications.argoproj.io guestbook -n argocd

# Clean up remaining resources
kubectl delete all --all -n argo-guestbook
```

**Note:** With `syncPolicy.automated.prune: true`, deleting the application will also delete managed resources.

### Common Status Checks

```bash
# Full status overview
kubectl get applications.argoproj.io guestbook -n argocd -o custom-columns=\
APPLICATION:.metadata.name,SYNC:.status.sync.status,HEALTH:.status.health.status,LAST_SYNC:.status.operationState.finishedAt

# Sync result details
kubectl get applications.argoproj.io guestbook -n argocd -o custom-columns=\
PHASE:.status.operationState.phase,MESSAGE:.status.operationState.message

# Resource-level status
kubectl get applications.argoproj.io guestbook -n argocd -o jsonpath='{.status.resources}' | jq .
```

## Troubleshooting

### Check Application Status

**For Tenants (via KubeSphere API):**

```bash
# Quick status check using labels
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/gitops.kubesphere.io/v1alpha1/namespaces/${DEVOPS_PROJECT}/applications/${APP_NAME}" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '{
  health: .metadata.labels["gitops.kubesphere.io/health-status"],
  sync: .metadata.labels["gitops.kubesphere.io/sync-status"],
  argocdApp: .metadata.labels["gitops.kubesphere.io/argocd-application"]
}'

# Detailed status from .status.argoApp (JSON string)
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/gitops.kubesphere.io/v1alpha1/namespaces/${DEVOPS_PROJECT}/applications/${APP_NAME}" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.status.argoApp' | jq -r '{
  syncStatus: .sync.status,
  healthStatus: .health.status,
  revision: .sync.revision,
  resources: [.resources[] | {kind: .kind, name: .name, status: .status, health: .health.status}],
  operationPhase: .operationState.phase,
  operationMessage: .operationState.message
}'
```

**For Admins (Direct ArgoCD Access):**

```bash
# Get application details
kubectl get application <app-name> -n argocd -o yaml

# Check application conditions
kubectl get application <app-name> -n argocd -o jsonpath='{.status.conditions}'

# View application events
kubectl describe application <app-name> -n argocd
```

### Common Issues

| Issue | Cause | Fix |
|-------|-------|-----|
| Sync failed | Invalid manifests | Check `status.operationState` for errors |
| Permission denied | RBAC issues | Verify ArgoCD has permissions in target namespace |
| Repo not found | Credential issues | Check repository secret and URL |
| Stuck in Progress | Resource stuck | Check resource health, may need manual intervention |
| OutOfSync | Drift detected | Enable auto-sync or manually sync |
| Resources not recreated after deletion | Auto-sync not enabled | Add `selfHeal: true` or manually trigger sync |
| Application status Unknown (KubeSphere GitOps) | Missing required label | Add label `gitops.kubesphere.io/argocd-location: argocd` |
| SharedResourceWarning / OutOfSync | Duplicate ArgoCD Applications | Delete manually-created ArgoCD Application, use only KubeSphere GitOps Application |

### Tips from Experience

**1. Use Full CRD Name in Commands:**
```bash
# Correct
kubectl get applications.argoproj.io guestbook -n argocd

# May fail (ambiguous)
kubectl get application guestbook -n argocd
```

**2. Automated Sync Behavior:**
- `prune: true` - Deletes resources not in Git
- `selfHeal: true` - Recreates resources deleted manually
- Both recommended for production GitOps workflows

**3. Namespace Creation:**
Use `CreateNamespace=true` sync option to auto-create target namespace:
```yaml
syncOptions:
  - CreateNamespace=true
```

**4. Force Refresh Timing:**
After adding refresh annotation, allow 5-10 seconds for reconciliation before checking status.

| Issue | Cause | Fix |
|-------|-------|-----|
| Sync failed | Invalid manifests | Check `status.operationState` for errors |
| Permission denied | RBAC issues | Verify ArgoCD has permissions in target namespace |
| Repo not found | Credential issues | Check repository secret and URL |
| Stuck in Progress | Resource stuck | Check resource health, may need manual intervention |
| OutOfSync | Drift detected | Enable auto-sync or manually sync |

### View Logs

```bash
# Application controller logs
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-application-controller

# Repo server logs
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-repo-server

# Server logs
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-server
```

### Debug Sync Failures

```bash
# Get detailed sync status
kubectl get application <app-name> -n argocd -o jsonpath='{.status.operationState}' | jq .

# Check resource health
kubectl get application <app-name> -n argocd -o jsonpath='{.status.resources}'

# Force refresh
argocd app get <app-name> --hard-refresh
```

## Integration with KubeSphere

### GitOps Applications Resource

KubeSphere provides a `applications.gitops.kubesphere.io` CRD that integrates with ArgoCD:

```yaml
apiVersion: gitops.kubesphere.io/v1alpha1
kind: Application
metadata:
  name: my-gitops-app
  namespace: my-project
spec:
  argoApp:
    source:
      repoURL: https://github.com/example/repo.git
      targetRevision: HEAD
      path: manifests
    destination:
      server: https://kubernetes.default.svc
      namespace: my-app
    syncPolicy:
      automated:
        prune: true
        selfHeal: true
```

### Multi-Cluster Deployments

ArgoCD can deploy to multiple clusters managed by KubeSphere:

```bash
# List registered clusters in ArgoCD
argocd cluster list

# Add a KubeSphere member cluster
argocd cluster add <kubeconfig-context-name>
```

## Version Information

| Component | Version |
|-----------|---------|
| ArgoCD | v2.11.7 |
| ArgoCD Helm Chart | 7.3.11 |
| Redis | 7.2.4 |
| Dex | v2.38.0 |


### Trigger Manual Sync via API

**Tenant Method (KubeSphere API):**
```bash
export API_TOKEN="<tenant-kubesphere-token>"

# Trigger sync
curl -s -X POST "${KUBESPHERE_API}/kapis/gitops.kubesphere.io/v1alpha1/namespaces/demo-project/applications/guestbook/sync" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"prune":true}'
```

**Notes:**
- Returns HTTP 400 with `another operation is already in progress` if auto-sync is enabled and recently completed
- The sync API is accessible to tenants
- Response may be empty on success

**Admin Method (ArgoCD API):**
```bash
# Get ArgoCD admin token
ARGO_TOKEN=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)

# Trigger sync
curl -s -X POST "https://argocd-server.argocd/api/v1/applications/guestbook/sync" \
  -H "Authorization: Bearer ${ARGO_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"prune":true,"dryRun":false}'
```
## References

- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [ArgoCD Helm Chart](https://github.com/argoproj/argo-helm/tree/main/charts/argo-cd)
- [KubeSphere GitOps Guide](https://docs.kubesphere.io/)
- [DevOps Extension README](/root/go/src/github.com/kubesphere/kse-extensions/devops/README.md)
