---
name: kubesphere-gateway
description: KubeSphere Gateway extension management Skill (ingress-nginx based, uses Kubernetes Ingress API + Gateway CRD gateway.kubesphere.io/v2alpha2). For the newer Kubernetes Gateway API (Traefik + GatewayProxy CRD), see the kubesphere-gateway-api skill instead. Covers installation, uninstallation, status checks, gateway status inspection, and troubleshooting (gateway stuck states, Helm failures, pod issues).
---

# KubeSphere Gateway

## Overview

Provides external access management (ingress) for KubeSphere using **ingress-nginx**. Supports three-tier gateway management:

| Tier | Scope | Name Pattern | Namespace | Label |
|---|---|---|---|
| **Cluster** | Entire cluster | `kubesphere-router-cluster` | `kubesphere-controls-system` | `kubesphere.io/gateway-type=cluster` |
| **Workspace** | Single workspace | `kubesphere-router-workspace-{workspace}` | `kubesphere-controls-system` | `kubesphere.io/gateway-type=workspace` |
| **Project** | Single project/namespace | `kubesphere-router-{namespace}` | `kubesphere-controls-system` | `kubesphere.io/gateway-type=project` |

Each gateway is a standalone Helm release of ingress-nginx. The gateway-controller-manager manages the lifecycle (install/upgrade/uninstall) via Helm.

### Core CRDs

- **`Gateway`** (`gateway.kubesphere.io/v2alpha2`) — represents a single ingress-nginx deployment. Key fields:
  - `spec.appVersion` — the Helm chart version (e.g. `kubesphere-nginx-ingress-<version>`)
  - `spec.values` — Helm values for ingress-nginx (controller config, service type, resources, etc.)
  - `status.state` — `Creating`, `Updating`, `Running`, `Faulted`, `Stopped`
  - `status.conditions[].type=GatewayReady` — True when fully operational
  - `status.loadBalancer` — LB ingress IPs/hostnames
  - `status.service` — Service type, ports, external IPs

- **`UpgradePlan`** (`gateway.kubesphere.io/v2alpha2`) — batch gateway upgrade job. Key fields:
  - `spec.gatewayReferences` — list of `{name, namespace}` to upgrade
  - `spec.targetAppVersion` — target version
  - `status.state` — `Pending`, `Running`, `Succeeded`, `Failed`

### Monitoring Integration

Gateway exposes NGINX metrics (requests, 4xx/5xx, latency P50/P90/P99) via Prometheus. Requires the `whizard-monitoring` extension (optional dependency).

---

## Before You Start

Check if Gateway extension is already installed:

```bash
kubectl get installplans.kubesphere.io gateway --ignore-not-found
```

If found, upgrading is supported — just select a newer version in Step 1.

---

## Installation

### Step 1: Detect and Select Version

```bash
ALL_VERSIONS=$(kubectl get extensionversions.kubesphere.io \
  -l kubesphere.io/extension-ref=gateway \
  -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V)

LATEST_STABLE=$(echo "$ALL_VERSIONS" | grep -v -E 'alpha|beta|rc' | tail -1)
if [ -z "$LATEST_STABLE" ]; then
  LATEST_STABLE=$(echo "$ALL_VERSIONS" | tail -1)
fi

echo "Available versions:"
echo "$ALL_VERSIONS"
echo ""
echo "Latest stable: $LATEST_STABLE"
```

This sets `ALL_VERSIONS` and `LATEST_STABLE`. Use `SELECTED_VERSION` for the version chosen.

Use the `question` tool:
- **`$LATEST_STABLE (Recommended)`** — accept the auto-detected version
- *(custom)* — type a specific version; validate it against the printed list

### Step 2: Detect and Select Clusters

```bash
CLUSTER_DATA=$(kubectl get clusters.cluster.kubesphere.io \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.conditions[?(@.type=="Ready")].status}{"\n"}{end}')

READY_CLUSTERS=$(echo "$CLUSTER_DATA" | awk -F'\t' '$2 == "True" {print $1}')
CLUSTER_COUNT=$(echo "$READY_CLUSTERS" | wc -l)

HOST_CLUSTER=$(kubectl get clusters.cluster.kubesphere.io \
  -l 'cluster-role.kubesphere.io/host' \
  -o jsonpath='{.items[0].metadata.name}' || echo "")

echo "Ready clusters:"
echo "$READY_CLUSTERS"
echo ""
echo "Cluster count: $CLUSTER_COUNT"
echo "Host cluster: $HOST_CLUSTER"
```

This sets `READY_CLUSTERS`, `CLUSTER_COUNT`, `HOST_CLUSTER`.

- **1 cluster** → skip selection, auto-use it. Set `TARGET_CLUSTERS="$HOST_CLUSTER"`
- **Multiple clusters** → use `question` with `multiple: true`:
  - **All clusters** → `TARGET_CLUSTERS="$READY_CLUSTERS"`
  - **Host cluster only** → `TARGET_CLUSTERS="$HOST_CLUSTER"`
  - *(custom)* — validate each name against `$READY_CLUSTERS`

### Step 3: Generate and Apply InstallPlan

```bash
./scripts/generate-installplan.sh "$SELECTED_VERSION" "$TARGET_CLUSTERS"
```

This generates the YAML to `/tmp/gateway-installplan.yaml`, runs `--dry-run=server`, then prints the apply command.

> For configurable extension values (ingress-nginx default settings, image registry, upgrade tool config, etc.), see [references/extension-values.md](references/extension-values.md).

Apply it:

```bash
kubectl apply -f /tmp/gateway-installplan.yaml
```

Tell the user "Installing". Then ask if they want to check status. If yes:

```bash
./scripts/check-status.sh poll
```

---

## Status Checking

| Purpose | Command |
|---|---|
| Single snapshot | `./scripts/check-status.sh quick` |
| Wait until complete (5min timeout) | `./scripts/check-status.sh poll` |

Logic:
- **All `Installed`** → ✓ success
- **Any `Failed`** → ✗ prints full status
- **Timeout (300s)** → ⚠ prints current status
- **In progress** → prints every 10s

---

## Uninstallation

> ⚠ **Always confirm with the user before proceeding.**

### Uninstall from all clusters

```bash
if ! kubectl get installplans.kubesphere.io gateway &>/dev/null; then
  echo "Gateway is not installed."
  exit 0
fi
```

Confirm with the user, then delete:

```bash
kubectl delete installplans.kubesphere.io gateway --ignore-not-found
```

Verify cleanup:

```bash
./scripts/verify-uninstall.sh
```

Success criteria:
1. InstallPlan is deleted
2. No active pods remain in `extension-gateway` namespace

### Uninstall from specific clusters

> **WARNING**: Do NOT delete the InstallPlan. Only remove target clusters from the placement list.

Confirm which clusters to remove, compute remaining clusters, then patch:

```bash
kubectl patch installplans.kubesphere.io gateway --type='json' \
  -p='[{"op": "replace", "path": "/spec/clusterScheduling/placement/clusters", "value": ["<REMAINING_CLUSTER_1>", "<REMAINING_CLUSTER_2>"]}]'
```

Success: patch returns OK + removed clusters no longer in `.status.clusterSchedulingStatuses`.

---

## Gateway Operations

### List Gateways

Gateways are organized by tier (see [Overview](#overview)), with each tier identified by the label `kubesphere.io/gateway-type`. List them grouped by tier:

```bash
echo "=== Cluster Gateway ==="
kubectl get gateways.gateway.kubesphere.io -n kubesphere-controls-system \
  -l kubesphere.io/gateway-type=cluster

echo -e "\n=== Workspace Gateways ==="
kubectl get gateways.gateway.kubesphere.io -n kubesphere-controls-system \
  -l kubesphere.io/gateway-type=workspace

echo -e "\n=== Project Gateways ==="
kubectl get gateways.gateway.kubesphere.io -n kubesphere-controls-system \
  -l kubesphere.io/gateway-type=project
```

### Check Gateway Status

Pick a gateway name from the [List Gateways](#list-gateways) output and run:

```bash
GW_NS="kubesphere-controls-system"
GW_NAME="<gateway-name-from-list>"

# app.kubernetes.io/instance uses the Helm release name if available, otherwise the Gateway name
GW_INSTANCE=$(kubectl get gateways.gateway.kubesphere.io -n $GW_NS $GW_NAME -o jsonpath='{.status.helmRelease.name}' 2>/dev/null)
if [ -z "$GW_INSTANCE" ]; then
  GW_INSTANCE="$GW_NAME"
fi

kubectl get gateways.gateway.kubesphere.io -n $GW_NS $GW_NAME -o wide
kubectl describe gateways.gateway.kubesphere.io -n $GW_NS $GW_NAME
kubectl get gateways.gateway.kubesphere.io -n $GW_NS $GW_NAME -o yaml
kubectl get pods -n $GW_NS -l "app.kubernetes.io/instance=$GW_INSTANCE"
```

Gateway states:
| State | Meaning |
|---|---|
| `Creating` | First-time Helm install in progress |
| `Updating` | Helm upgrade in progress (spec changed) |
| `Running` | Fully operational (all replicas available) |
| `Faulted` | Deployment missing, stopped unexpectedly, or health probe timeout |
| `Stopped` | Scaled to zero replicas intentionally |

---

## Troubleshooting

> Set `$GW_NAME` according to the gateway tier being troubleshot (see naming rules in [Overview](#overview)):
> - Cluster → `GW_NAME=kubesphere-router-cluster`
> - Workspace → `GW_NAME=kubesphere-router-workspace-${WORKSPACE}`
> - Project → `GW_NAME=kubesphere-router-${NAMESPACE}`
>
> Common namespace: `GW_NS=kubesphere-controls-system`
>
> `$GW_INSTANCE` is auto-resolved from `status.helmRelease.name` (falls back to `$GW_NAME`). If not yet set, run:
> ```bash
> GW_INSTANCE=$(kubectl get gateways.gateway.kubesphere.io -n $GW_NS $GW_NAME -o jsonpath='{.status.helmRelease.name}' 2>/dev/null)
> if [ -z "$GW_INSTANCE" ]; then
>   GW_INSTANCE="$GW_NAME"
> fi
> ```

### Gateway stuck in `Creating` or `Updating` state

```bash
kubectl describe gateways.gateway.kubesphere.io -n $GW_NS $GW_NAME

kubectl logs -n extension-gateway -l app=gateway-controller-manager --tail=200 | grep -iE "(error|helm|install|upgrade|reconcile)"

kubectl get deployment -n $GW_NS -l "app.kubernetes.io/instance=$GW_INSTANCE,app.kubernetes.io/component=controller"

kubectl get configmap -n $GW_NS $GW_NAME -o yaml
```

Common causes: Chart ConfigMap missing/corrupted, Helm wrapper timeout, invalid `spec.values`.

### Gateway shows `Faulted` state

```bash
kubectl get deployment -n $GW_NS $GW_NAME -o wide
kubectl describe deployment -n $GW_NS $GW_NAME
kubectl get pods -n $GW_NS -l "app.kubernetes.io/instance=$GW_INSTANCE" -o wide

POD_NAME=$(kubectl get pods -n $GW_NS -l "app.kubernetes.io/instance=$GW_INSTANCE" -o jsonpath='{.items[0].metadata.name}')
kubectl describe pod -n $GW_NS $POD_NAME
kubectl logs -n $GW_NS $POD_NAME --tail=100
```

Common causes: Image pull failure, resource constraints, port conflicts, missing ConfigMap/Secret.

### Gateway pod crash-looping / CrashLoopBackOff

```bash
kubectl logs -n $GW_NS -l "app.kubernetes.io/instance=$GW_INSTANCE" --tail=100 --previous
kubectl get events -n $GW_NS --sort-by='.lastTimestamp' | tail -20
kubectl exec -n $GW_NS -l "app.kubernetes.io/instance=$GW_INSTANCE" -- cat /etc/nginx/nginx.conf 2>/dev/null | head -50
kubectl get configmap -n $GW_NS -l "app.kubernetes.io/instance=$GW_INSTANCE" -o yaml
```

Common causes: Misconfigured nginx config, port conflicts, resource limits (OOMKilled), missing dependencies (ConfigMap/Secret).

### Log search not working

Gateway log search proxies to `whizard-telemetry-apiserver`:

```bash
kubectl get configmap -n extension-gateway gateway-agent-backend-config -o yaml
kubectl get pods -n extension-whizard-telemetry
kubectl get svc -n extension-whizard-telemetry whizard-telemetry-apiserver
kubectl logs -n extension-gateway -l app=gateway-apiserver --tail=100 | grep -iE "(log|search|whizard|proxy)"
```

Common causes: Whizard-telemetry not installed or not running, misconfigured `gateway-agent-backend-config`, network policy blocking cross-namespace traffic.

### Gateway controller not reconciling

```bash
kubectl get pods -n extension-gateway -l app=gateway-controller-manager
kubectl logs -n extension-gateway -l app=gateway-controller-manager --tail=200
kubectl get validatingwebhookconfiguration -l "app.kubernetes.io/managed-by=Helm,kubesphere.io/extension-ref=gateway"
kubectl get deployment -n extension-gateway -l app=gateway-controller-manager -o yaml
```

Common causes: Controller pod not running, webhook configuration blocking updates, Helm release state mismatch, RBAC permission issues.
```