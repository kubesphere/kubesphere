---
name: kubeeye
description: Use when deploying KubeEye for cluster inspection, creating InspectRule/InspectPlan resources, or retrieving inspection results. Covers InstallPlan-based deployment, OPA/PromQL/FileChange/Sysctl/Systemd/NodeInfo/FileFilter/ServiceConnect/CustomCommand rule types, and report retrieval. Always consult this skill when the user mentions KubeEye, cluster inspection, InspectRule, InspectPlan, or inspection reports.
compatibility: Requires kubectl and a KubeSphere cluster with the kubeeye extension published in the marketplace.
---

# KubeEye

## Overview

KubeEye is a Kubernetes cluster inspection tool for KubeSphere. It detects issues in workloads, nodes, configurations, and components through OPA/Rego policies, PromQL queries, file integrity checks, kernel parameter validation, and systemd health checks. It is installed as a KubeSphere extension.

## When to Use

- Installing or configuring the KubeEye cluster inspection extension in KubeSphere
- Writing cluster inspection rules (OPA, PromQL, file checks, etc.)
- Configuring periodic or one-shot inspection tasks
- Viewing or downloading cluster inspection reports

## Architecture

### Extension Installation Flow

```
kspublish (push extension to KubeSphere)
    |
    v
Extension available in KubeSphere marketplace
    |
    v
kubectl apply -f installplan.yaml
    |
    v
KubeEye deployed (3 components)
```

### Runtime Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     KubeSphere Extension                    │
│                                                             │
│  ┌─────────────────────┐   ┌─────────────────────────────┐  │
│  │   kubeeye-apiserver │   │  kubeeye-controller-manager │  │
│  │   (Gin REST API)    │   │  (4 CRD Controllers)        │  │
│  │   Port 9090         │   │                             │  │
│  └──────────┬──────────┘   └──────────────┬──────────────┘  │
│             │                              │                 │
│             ▼                              ▼                 │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                    CRDs                              │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │   │
│  │  │InspectRule│ │InspectPlan│ │InspectTask│ │Inspect │ │   │
│  │  │          │ │          │ │          │ │Result  │ │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └────────┘ │   │
│  └──────────────────────────────────────────────────────┘   │
│                              │                              │
│                              ▼                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              kubeeye-job (K8s Jobs)                  │   │
│  │  OPA ├─ PromQL ├─ FileChange ├─ Sysctl ├─ Systemd   │   │
│  │  NodeInfo ├─ FileFilter ├─ ServiceConnect ├─ Cmd    │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

> **Note:** `Cmd` in the diagram refers to the `customCommand` rule type.

### CRD Overview

| CRD | API Version | Scope | Purpose |
|-----|-------------|-------|---------|
| **InspectRule** | `kubeeye.kubesphere.io/v1alpha2` | Cluster | Defines inspection rules (OPA, PromQL, file checks, etc.) |
| **InspectPlan** | `kubeeye.kubesphere.io/v1alpha2` | Cluster | Schedules inspection execution (cron or one-shot) |
| **InspectTask** | `kubeeye.kubesphere.io/v1alpha2` | Cluster | Tracks a single inspection execution (created by InspectPlan) |
| **InspectResult** | `kubeeye.kubesphere.io/v1alpha2` | Cluster | Stores inspection results (populated by InspectTask) |

### CRD Data Flow

```
User creates InspectRule ──────┐
                               ├──> InspectPlan references InspectRules
User creates InspectPlan ──────┘         |
                                         | (cron trigger or manual)
                                         v
                               InspectTask created by InspectPlanReconciler
                                         |
                                    InspectTaskReconciler:
                                    1. Fetches referenced InspectRules
                                    2. Merges rules per type
                                    3. Creates K8s Jobs (kubeeye-job)
                                    4. Jobs execute rule checks
                                    5. Results accumulated
                                         |
                                         v
                               InspectResult populated with findings
                                         |
                                         v
                               User views results via:
                               - kubectl get inspectresult
                               - API: HTML report / XLSX download
```

## Before You Start

Check if the KubeEye extension is available:

```bash
kubectl get extensionversions | grep kubeeye
```

Expected output: `kubeeye-{version}` (e.g. `kubeeye-1.0.1`).

If nothing is shown, the extension hasn't been published to KubeSphere yet.

Check if KubeEye is already installed:

```bash
kubectl get installplans.kubesphere.io kubeeye --ignore-not-found
```

If the InstallPlan exists, upgrading is supported — just select a newer version.

## Installation

KubeEye is installed as a KubeSphere extension. The extension must first be published to KubeSphere (via `kspublish`, assumed to be already available).

> **Note:** The `./scripts/` paths below assume you are running from the `skills/kubeeye/` directory. Adjust the path if you are running from elsewhere.

### Step 1: Detect and Select Version

```bash
ALL_VERSIONS=$(kubectl get extensionversions.kubesphere.io \
  -l kubesphere.io/extension-ref=kubeeye \
  -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V)

LATEST_STABLE=$(echo "$ALL_VERSIONS" | tail -1)

echo "Available versions:"
echo "$ALL_VERSIONS"
echo ""
echo "Latest stable: $LATEST_STABLE"
```

This sets `ALL_VERSIONS` and `LATEST_STABLE`. Use `SELECTED_VERSION` for the version chosen.

### Step 2: Generate and Apply InstallPlan

```bash
./scripts/generate-installplan.sh "$SELECTED_VERSION"
```

This generates the YAML to `/tmp/kubeeye-installplan.yaml`, runs `--dry-run=server`, then prints the apply command.

Apply it:

```bash
kubectl apply -f /tmp/kubeeye-installplan.yaml
```

Tell the user "Installing". Then ask if they want to check status. If yes:

```bash
./scripts/check-status.sh poll
```

## Quick Start

### Step 1: Import Rules

```bash
kubectl apply -f rules/
```

This applies all sample InspectRule files bundled in this skill's `rules/` directory.

> **Note**: Set the correct `prometheus.endpoint` in `kubeeye_promql_inspect.yaml` before importing if using PromQL rules.

### Step 2: Create InspectPlan

```bash
./scripts/generate-plan.sh
```

This creates an InspectPlan referencing all available InspectRules and applies it directly.

### Step 3: Monitor InspectTask

Once an InspectPlan is applied, the controller creates an InspectTask:

```bash
# Watch task status
kubectl get inspecttask -w

# Describe a task
kubectl describe inspecttask {task-name}
```

### Step 4: View InspectResult

When the InspectTask completes, an InspectResult is created:

```bash
# List results
kubectl get inspectresult

# View result details
kubectl get inspectresult {result-name} -o yaml

# Download HTML report
kubectl get svc -n extension-kubeeye kubeeye-apiserver \
  -o custom-columns=CLUSTER-IP:.spec.clusterIP,PORT:.spec.ports[*].port
curl http://{svc-ip}:9090/kapis/kubeeye.kubesphere.io/v1alpha2/inspectresults/{result-name}?type=html -o report.html

# Download XLSX report
curl http://{svc-ip}:9090/kapis/kubeeye.kubesphere.io/v1alpha2/inspectresults/{result-name}/download -o report.xlsx
```

### View in Browser

```bash
kubectl -n extension-kubeeye expose deploy kubeeye-apiserver --port=9090 --type=NodePort --name=ke-apiserver-node-port
# http://{node-address}:{node-port}/kapis/kubeeye.kubesphere.io/v1alpha2/inspectresults/{result-name}?type=html
```

## Status Checking

| Purpose | Command |
|---------|---------|
| Single snapshot | `./scripts/check-status.sh quick` |
| Wait until complete (5min timeout) | `./scripts/check-status.sh poll` |

Logic:
- **All `Installed`** → ✓ success
- **Any `Failed`** → ✗ prints full status
- **Timeout (300s)** → ⚠ prints current status
- **In progress** → prints every 10s

## InspectRule Development

An InspectRule (`kubeeye.kubesphere.io/v1alpha2`) defines the inspection logic. A single rule file can combine multiple rule types.

### OPA Rules

Uses Rego policy language. Specify `input.kind` and `input.apiVersion`.

**Sub-packages:**
- `inspect.kubeeye` — Standard K8s resources (Deployment, Pod, Node, ConfigMap, etc.)
- `inspect.kubeeye.nodeStatsSummary` — Node stats summary (`input.pods[*]` with ephemeral-storage)

**Example - Deployment imagePullPolicy check:**

```yaml
spec:
  opas:
  - name: imagePullPolicyRule
    rule: |-
      package inspect.kubeeye
      import rego.v1
      deny contains msg if {
        input.kind == "Deployment"
        input.apiVersion == "apps/v1"
        container := input.spec.template.spec.containers[_]
        container.imagePullPolicy != "IfNotPresent"
        msg := {
            "Name": input.metadata.name,
            "Namespace": input.metadata.namespace,
            "Type": input.kind,
            "Message": "ImagePullPolicyNotIfNotPresent",
            "Reason": sprintf("imagePullPolicy is %v, should be IfNotPresent", [container.imagePullPolicy]),
            "Level": "WARNING"
        }
      }
```

**Example - Node ephemeral-storage check:**

```yaml
spec:
  opas:
  - name: CheckEphemeralStorage
    rule: |-
      package inspect.kubeeye.nodeStatsSummary
      import rego.v1
      threshold := 5 * 1024 * 1024 * 1024
      deny contains msg if {
        pod := input.pods[_]
        bytes := pod["ephemeral-storage"].usedBytes
        bytes > threshold
        msg := {
            "Name": pod.podRef.name,
            "Namespace": pod.podRef.namespace,
            "Type": "Pod",
            "Level": "danger",
            "Message": sprintf("ephemeral-storage usage %.2f GB exceeds 5 GB", [bytes / 1073741824]),
            "Reason": "ephemeral-storage exceeds threshold"
        }
      }
```

### PromQL Rules

```yaml
spec:
  prometheus:
    endpoint: http://prometheus-k8s.monitoring.svc.cluster.local:9090
  promQL:
  - name: NodeMemory
    desc: Node memory usage > 30%
    rule: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 30
    rawDataEnabled: true
```

### File Change Rules

```yaml
spec:
  fileChange:
  - name: kubelet-config
    path: /var/lib/kubelet/config.yaml
    level: warning
```

### Sysctl Rules

```yaml
spec:
  sysctl:
  - name: net.ipv4.ip_forward
    rule: net.ipv4.ip_forward = 1
    level: warning
```

### Systemd Rules

```yaml
spec:
  systemd:
  - name: kubelet
    rule: kubelet == "active"
    level: warning
```

### Node Info Rules

Supported `resourcesType`: `cpu`, `memory`, `filesystem`, `inode`, `load`.

```yaml
spec:
  nodeInfo:
  - name: CpuUsage
    rule: cpu > 20
    resourcesType: cpu
    desc: CPU usage > 20%
    level: warning
```

### File Content Filter Rules

```yaml
spec:
  fileFilter:
  - name: systemLog
    path: /var/log/syslog
    rule: error
    level: warning
```

### Service Connectivity Rules

```yaml
spec:
  serviceConnect:
  - workspace: system-workspace
    level: warning
```

### Custom Command Rules

```yaml
spec:
  customCommand:
  - name: check-disk
    command: "df -h / | tail -1"
    rule: ".*5[0-9]%"
    level: warning
```

### Component Exclude

```yaml
spec:
  componentExclude:
  - "kube-system/kube-dns"
```

## InspectPlan Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `schedule` | string | Cron expression (e.g. `"*/30 * * * ?"`). Remove for one-shot. |
| `suspend` | bool | Pause periodic inspection |
| `timeout` | string | Inspection timeout (default: `"10m"`) |
| `ruleNames` | array | List of InspectRule names. Supports `nodeName`/`nodeSelector` per rule. |
| `maxTasks` | int | Max retained results (older ones cleaned up) |
| `once` | timestamp | One-shot inspection at a specific time |
| `clusterName` | array | Multi-cluster targets (KubeSphere multi-cluster) |

## Operations

### View Logs

```bash
kubectl logs -n extension-kubeeye -l control-plane=controller-manager --tail=100
kubectl logs -n extension-kubeeye -l app=kubeeye-apiserver --tail=100
```

### Email Notification

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: message-secret
  namespace: extension-kubeeye
type: Opaque
stringData:
  username: your-email@example.com
  password: your-password
```

Update ConfigMap `kubeeye-config`:

```yaml
data:
  config: |-
    job:
      autoDelTime: 30
      backLimit: 5
      image: kubespheredev/kubeeye-job:v1.0.6
      imagePullPolicy: Always
      resources:
        limits:
          cpu: 2000m
          memory: 512Mi
        requests:
          cpu: 50m
          memory: 256Mi
    message:
      enable: true
      email:
        address: smtp.example.com
        port: 25
        fo: sender@example.com
        to:
        - recipient@example.com
        secretKey: message-secret
```

## Uninstallation

> ⚠ **Always confirm with the user before proceeding.**

```bash
if ! kubectl get installplans.kubesphere.io kubeeye --ignore-not-found &>/dev/null; then
  echo "KubeEye is not installed."
  exit 0
fi
```

Confirm with the user, then delete:

```bash
kubectl delete installplans.kubesphere.io kubeeye --ignore-not-found
```

Verify cleanup:

```bash
./scripts/verify-uninstall.sh
```

Success criteria:
1. InstallPlan is deleted
2. No pods remain in `extension-kubeeye`

## Troubleshooting

### Pods Not Starting

```bash
kubectl describe po -n extension-kubeeye
```

### Inspection Not Running

```bash
kubectl get inspectplan
kubectl get inspectrule
kubectl describe inspectplan inspectplan
```

### No Results

```bash
kubectl get inspecttask
kubectl get inspectresult
kubectl get endpoints -n extension-kubeeye kubeeye-apiserver
```
