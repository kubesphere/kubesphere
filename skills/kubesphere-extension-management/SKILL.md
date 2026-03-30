---
name: kubesphere-extension-management
description: KubeSphere extension management Skill. Use when user requests to install, configure, upgrade, uninstall extensions, or query extension info/troubleshoot issues. Includes extension discovery, dependency management, install configuration, version management.
---

# KubeSphere Extension Management


---

## Architecture


```
┌──────────────────────┐        ┌─────────────────────────────┐        ┌──────────────────────┐        ┌──────────────────────┐
│  Extension Museum    │        │  Extension                  │        │  InstallPlan         │        │  Deployed Extension  │
│  (Local Chart Repo)  │───────▶│  - description              │───────▶│  - extension         │───────▶│                      │
│                      │ (sync) │  - status                   │        │  - config            │        │  Namespace: <target> │
└──────────────────────┘        │  ExtensionVersion           │        │  - clusterScheduling │        │  └── Pods, Services   │
                                │  - version                  │        └──────────────────────┘        └──────────────────────┘
                                │  - chartURL                 │              │ (reconcile)
                                │  - externalDependencies     │              |
                                │  - installationMode         │              |
                                └─────────────────────────────┘              ▼
                                                                    ┌──────────────────────┐
                                                                    │  Job                 │
                                                                    │  helm-upgrade-<name> │
                                                                    │  - helm install/     │
                                                                    │    upgrade/uninstall │
                                                                    └──────────┬───────────┘
                                                                               │
                                                                               ▼
                                                                    ┌──────────────────────┐
                                                                    │  Pod                 │
                                                                    │  (Helm execution)    │
                                                                    └──────────────────────┘
```

## Core Concepts

### Extension
Metadata and status for available/installed extensions.

| Field | Description | Example |
|-------|-------------|---------|
| `metadata.name` | Extension name | `whizard-monitoring` |
| `spec.versions[]` | Available versions | `[1.2.0, 1.2.1]` |
| `status.state` | Current state | `Installed`/`Upgrading`/`Failed` |
| `status.enabled` | Enabled status | `true`/`false` |
| `status.installedVersion` | Installed version | `1.2.1` |

---

### ExtensionVersion
Version-specific details: Helm chart location, dependencies, requirements.

| Field | Description | Example |
|-------|-------------|---------|
| `metadata.name` | Version resource name | `whizard-monitoring-1.2.1` |
| `spec.version` | Version number | `1.2.1` |
| `spec.chartURL` | Helm chart URL | `https://extensions-museum.../whizard-monitoring-1.2.1.tgz` |
| `spec.namespace` | Target namespace | `kubesphere-monitoring-system` |
| `spec.installationMode` | Install mode | `Multicluster`/`HostOnly` |
| `spec.ksVersion` | Required KS version | `>=4.2.0-0` |
| `spec.externalDependencies[]` | Required extensions | `[name: whizard-telemetry]` |

---

### InstallPlan
Trigger for install/upgrade/uninstall. Creates Helm Job to deploy components.

**Spec**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `metadata.name` | string | ✅ | Must match `spec.extension.name` |
| `spec.extension.name` | string | ✅ | Extension name |
| `spec.extension.version` | string | ✅ | Exact version to install |
| `spec.enabled` | bool | ✅ | Enable extension |
| `spec.upgradeStrategy` | string | ✅ | Use `Manual` for production |
| `spec.config` | string | ❌ | Custom YAML config |
| `spec.clusterScheduling` | object | ❌ | Multi-cluster config (Multicluster mode only) |

**Status**:
| Field | Description |
|-------|-------------|
| `status.state` | `Installed`/`Installing`/`Upgrading`/`Failed` |
| `status.jobName` | Helm upgrade Job name |
| `status.targetNamespace` | Target namespace |
| `status.conditions[]` | Status conditions with messages |

**⚠️ CRITICAL:**
- `metadata.name` = `spec.extension.name`
- Use EXACT version from user request
- Set `enabled: true` and `upgradeStrategy: Manual`



## Step-by-Step Workflow

### 1. List and Inspect Extensions

```bash
# List all extensions
kubectl get extensions

# List by category
for c in $(kubectl get categories.kubesphere.io -o jsonpath='{.items[*].metadata.name}'); do
 exts=$(kubectl get extensions.kubesphere.io -l kubesphere.io/category="$c" -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')
 [ -n "$exts" ] && echo -e "$c:\n$exts"
done

# List extension available versions
kubectl get extensionversions.kubesphere.io -l kubesphere.io/extension-ref=<extension-name>

# View extension details
kubectl describe extension <extension-name>
kubectl describe extensionversion.kubesphere.io <extension-name>-<version>
```

### 2. Install Extension

**⚠️ CRITICAL:**
- InstallPlan `metadata.name` MUST match `spec.extension.name`
- Use EXACT version from user request (not `recommendedVersion` or `latest`)

#### 2.1 Verify Version Exists

```bash
# Verify the extension and version exist before creating InstallPlan
kubectl get extension <extension-name>
kubectl get extensionversion <extension-name>-<version>

# Get version details (check installationMode, namespace, dependencies, etc.)
kubectl describe extension <extension-name>
kubectl describe extensionversion <extension-name>-<version>
```

#### 2.2 Create InstallPlan

Based on the extension details above:
- **config**: Omit if user didn't request custom configuration
- **clusterScheduling**: Only include if `installationMode=Multicluster`, specify clusters in `placement.clusters`


```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: <extension-name>
spec:
  # clusterScheduling:    # Only for Multicluster extensions.
  #   placement:
  #     clusters:
  #     - <cluster-name>
  #   overrides:
  #     host: |-
  #       # Extension agent configuration: custom settings for the extension agent in the current cluster, with higher priority than the global extension configuration.
  # 
  #       custom:
  #         key: override-value
  # config: |  ## Omit if user didn't request custom config
  #   # Custom configuration for the extension, serving as a global extension configuration that can override default settings.
  # 
  #   custom:
  #     key: override-value
  enabled: true
  extension:
    name: <extension-name>     # CRITICAL: MUST match metadata.name
    version: <version>         # CRITICAL: Use EXACT version requested
  upgradeStrategy: Manual
```

```bash
kubectl apply -f installplan-<extension-name>.yaml
```

### 3. Track & Verify Installation

```bash
# Watch status
kubectl get installplan <extension-name> -w

# Get details
kubectl describe installplan <extension-name>

# Verify extension status
kubectl describe extension <extension-name>

# Check deployed resources
NAMESPACE=$(kubectl get installplan <extension-name> -o jsonpath='{.status.targetNamespace}')
kubectl get pods,svc -n $NAMESPACE
```

### 4. Update & Upgrade Extension

**⚠️ CRITICAL: InstallPlan `metadata.name` MUST match `spec.extension.name`**

- **Update**: Modify `spec.config` or `spec.clusterScheduling` (version unchanged)
- **Upgrade**: Modify `spec.extension.version` to new version

```bash
# Get current InstallPlan as template
kubectl get installplan <extension-name> -o yaml > installplan-<extension-name>.yaml

# Edit and apply:
# - For Update: modify spec.config or spec.clusterScheduling
# - For Upgrade: change spec.extension.version to target version
kubectl apply -f installplan-<extension-name>.yaml
kubectl get installplan <extension-name> -w
```

### 5. Uninstall Extension

```bash
kubectl delete installplan <extension-name>
```

## Best Practices

- Use default config (omit `spec.config`) unless custom values are needed
- Always use `upgradeStrategy: Manual` for production
- Use exact version (not `recommendedVersion` or `latest`)
- Test in staging before production
- Review changelogs before upgrades
- Document custom configurations

## Troubleshooting

### Issue 1: InstallPlan Failed (Job Execution Failed)

**Symptoms:** InstallPlan stuck in "Installing"/"Upgrading", Job pod shows Error/Failed

**Diagnosis:**

```bash
# Step 1: Check InstallPlan status
kubectl describe installplan <extension-name>
# Check status.state and status.conditions

# Step 2: Check Job pod logs
NAMESPACE=$(kubectl get installplan <extension-name> -o jsonpath='{.status.targetNamespace}')
JOB_NAME=$(kubectl get installplan <extension-name> -o jsonpath='{.status.jobName}')
kubectl logs -n $NAMESPACE -l job-name=$JOB_NAME --tail=100

# Step 3: For Multicluster - check cluster scheduling status
kubectl get installplan <extension-name> -o jsonpath='{.status.clusterSchedulingStatuses}' | jq .
# Check each cluster's state and conditions, then get job name and logs
```

**Solutions:**
- Verify extension version exists
- Check dependencies are installed
- Check config syntax

---

### Issue 2: InstallPlan Status Not Updated (Job Completed)

**Symptoms:** Job pod completed successfully, but InstallPlan stays in "Installing"/"Upgrading"

**Diagnosis:**

```bash
# Check if job completed
NAMESPACE=$(kubectl get installplan <extension-name> -o jsonpath='{.status.targetNamespace}')
JOB_NAME=$(kubectl get installplan <extension-name> -o jsonpath='{.status.jobName}')
kubectl get pods -n $NAMESPACE -l job-name=$JOB_NAME

# Check job completion time vs current time
kubectl get job $JOB_NAME -n $NAMESPACE -o jsonpath='{.status.completionTime}'
```

**Cause:** Clock skew between nodes (NTP not synchronized)

**Solution:** Check and synchronize NTP across all cluster nodes

## Quick Reference

| Action | Command |
|--------|---------|
| List all extensions | `kubectl get extensions` |
| List extension versions | `kubectl get extensionversions.kubesphere.io -l kubesphere.io/extension-ref=<name>` |
| View extension details | `kubectl describe extension <name>` |
| View version details | `kubectl describe extensionversion <name>-<version>` |
| Install extension | `kubectl apply -f installplan-<name>.yaml` |
| Track installation | `kubectl get installplan <name> -w` |
| Upgrade extension | Modify version in InstallPlan, then `kubectl apply` |
| Uninstall extension | `kubectl delete installplan <name>` |

## References

- [KubeSphere Extensions Documentation](https://docs.kubesphere.com.cn/v4.2.1/06-extension-management/)
- [KubeSphere Extensions Marketplace](https://kubesphere.com.cn/marketplace/)

## Related Skills

- `kubesphere-core` - Core platform architecture
- `kubesphere-cluster-management` - Cluster operations
