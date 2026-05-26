---
name: kubesphere-gateway-api
description: KubeSphere Gateway API extension management Skill (Traefik based, uses Kubernetes Gateway API + GatewayProxy CRD gatewayapi.kubesphere.io/v1alpha1). This is the newer Kubernetes Gateway API standard. For the older Ingress API based gateway (ingress-nginx + Gateway CRD gateway.kubesphere.io/v2alpha2), see the kubesphere-gateway skill instead. Covers installation, uninstallation, status checks, GatewayProxy status inspection, and troubleshooting.
---
# KubeSphere Gateway API

## Overview

Provides external access management using **Kubernetes Gateway API** with **Traefik** as the underlying proxy implementation. Supports three-tier gatewayproxy management:

| Tier                | Scope                    | Name Pattern                           | Namespace                      | Description                   |
| ------------------- | ------------------------ | -------------------------------------- | ------------------------------ | ----------------------------- |
| **Cluster**   | Entire cluster           | `gatewayproxy-cluster`               | `kubesphere-controls-system` | Cluster-scoped GatewayProxy   |
| **Workspace** | Single workspace         | `gatewayproxy-workspace-{workspace}` | `kubesphere-controls-system` | Workspace-scoped GatewayProxy |
| **Project**   | Single project/namespace | `gatewayproxy-namespace-{namespace}` | `kubesphere-controls-system` | Namespace-scoped GatewayProxy |

Each **GatewayProxy** (`gatewayproxies.gatewayapi.kubesphere.io`) deploys a Traefik instance (the proxy implementation). It auto-creates a `GatewayClass`, and users can then create standard **Gateway** (`gateways.gateway.networking.k8s.io`) resources that reference that GatewayClass. The extension consists of three components:

- **backend-extension** — API server on the host cluster
- **backend-agent** — API server + controller-manager on every cluster (including the host cluster if selected)
- **frontend** — React SPA served via Nginx

### Key Differentiator from `kubesphere-gateway`

| Aspect               | kubesphere-gateway                                      | kubesphere-gateway-api                                                    |
| -------------------- | ------------------------------------------------------- | ------------------------------------------------------------------------- |
| Underlying proxy     | ingress-nginx                                           | Traefik                                                                   |
| API standard         | Custom Gateway CRD (`gateway.kubesphere.io/v2alpha2`) | Kubernetes Gateway API (`gateway.networking.k8s.io`) + GatewayProxy CRD |
| Core resource        | `Gateway` + standard `IngressClass`/`Ingress`     | `GatewayProxy` + standard `Gateway`/`GatewayClass`/`HTTPRoute`    |
| Lifecycle management | Helm release per gateway                                | Helm release per GatewayProxy                                             |

### Core CRDs

- **`GatewayProxy`** (`gatewayapi.kubesphere.io/v1alpha1`) — the proxy implementation (e.g. Traefik). When created, the controller deploys Traefik via Helm SDK and auto-creates a `GatewayClass`. Key fields:

  - `spec.type` — proxy type (currently only `Traefik`)
  - `spec.traefik.rawValues` — raw Helm values passed to the Traefik chart
  - `spec.traefik.deployment.replicas` — replica count
  - `spec.traefik.service.type` — Service type (ClusterIP, NodePort, LoadBalancer)
  - `spec.traefik.createDefaultGateway` — whether to auto-create a default Gateway
  - `spec.paused` — pause reconciliation
  - `status.conditions` — condition types: `Ready`, `Progressing`, `NewVersionDetected`
  - `status.service` — Service type, ports, external IPs, load balancer status
  - `status.entrypoints` — exposed entrypoints with ports and protocols
  - `status.gatewayClass.name` — auto-created GatewayClass name
  - `status.helmRelease.name` — Helm release name
- **`GatewayClass`** (`gateway.networking.k8s.io/v1`) — standard Kubernetes Gateway API class, auto-created by the GatewayProxy controller
- **`Gateway`** (`gateway.networking.k8s.io/v1`) — standard Kubernetes Gateway API gateway, associated with a GatewayClass
- **`HTTPRoute`** / **`GRPCRoute`** / **`TLSRoute`** / **`TCPRoute`** / **`UDPRoute`** — standard Kubernetes Gateway API route resources

### Multi-Tenant Scoping

GatewayProxy and Gateway are scoped via labels:

- `gatewayapi.kubesphere.io/scope-type` — `cluster`, `workspace`, or `namespace`
- `gatewayapi.kubesphere.io/scope-workspace` — workspace name (for workspace scope)
- `gatewayapi.kubesphere.io/scope-namespace` — namespace name (for namespace scope)

### Monitoring Integration

GatewayProxy exposes Traefik metrics via Prometheus. Requires the `whizard-monitoring` extension (optional dependency). Log search requires the `whizard-logging` extension. 

---

## Before You Start

Check if Gateway API extension is already installed:

```bash
kubectl get installplans.kubesphere.io gateway-api --ignore-not-found
```

If found, upgrading is supported — just select a newer version in Step 1.

---

## Installation

### Step 1: Detect and Select Version

```bash
ALL_VERSIONS=$(kubectl get extensionversions.kubesphere.io \
  -l kubesphere.io/extension-ref=gateway-api \
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

This generates the YAML to `/tmp/gateway-api-installplan.yaml`, runs `--dry-run=server`, then prints the apply command.

Apply it:

```bash
kubectl apply -f /tmp/gateway-api-installplan.yaml
```

Tell the user "Installing". Then ask if they want to check status. If yes:

```bash
./scripts/check-status.sh poll
```

---

## Status Checking

| Purpose                            | Command                             |
| ---------------------------------- | ----------------------------------- |
| Single snapshot                    | `./scripts/check-status.sh quick` |
| Wait until complete (5min timeout) | `./scripts/check-status.sh poll`  |

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
if ! kubectl get installplans.kubesphere.io gateway-api &>/dev/null; then
  echo "Gateway API is not installed."
  exit 0
fi
```

Confirm with the user, then delete:

```bash
kubectl delete installplans.kubesphere.io gateway-api --ignore-not-found
```

Verify cleanup:

```bash
./scripts/verify-uninstall.sh
```

Success criteria:

1. InstallPlan is deleted
2. No active pods remain in `extension-gateway-api` namespace

### Uninstall from specific clusters

> **WARNING**: Do NOT delete the InstallPlan. Only remove target clusters from the placement list.

Confirm which clusters to remove, compute remaining clusters, then patch:

```bash
kubectl patch installplans.kubesphere.io gateway-api --type='json' \
  -p='[{"op": "replace", "path": "/spec/clusterScheduling/placement/clusters", "value": ["<REMAINING_CLUSTER_1>", "<REMAINING_CLUSTER_2>"]}]'
```

Success: patch returns OK + removed clusters no longer in `.status.clusterSchedulingStatuses`.

---

## GatewayProxy Operations

### List GatewayProxies

GatewayProxies are organized by scope type. List them:

```bash
echo "=== Cluster GatewayProxies ==="
kubectl get gatewayproxies.gatewayapi.kubesphere.io -A \
  -l gatewayapi.kubesphere.io/scope-type=cluster

echo -e "\n=== Workspace GatewayProxies ==="
kubectl get gatewayproxies.gatewayapi.kubesphere.io -A \
  -l gatewayapi.kubesphere.io/scope-type=workspace

echo -e "\n=== Namespace GatewayProxies ==="
kubectl get gatewayproxies.gatewayapi.kubesphere.io -A \
  -l gatewayapi.kubesphere.io/scope-type=namespace
```

### Check GatewayProxy Status

Pick a gateway proxy name from the list above and run:

```bash
GWP_NS="kubesphere-controls-system"
GWP_NAME="<gatewayproxy-name-from-list>"

# app.kubernetes.io/instance uses the Helm release name if available, otherwise the GatewayProxy name
GWP_INSTANCE=$(kubectl get gatewayproxies.gatewayapi.kubesphere.io -n $GWP_NS $GWP_NAME -o jsonpath='{.status.helmRelease.name}' 2>/dev/null)
if [ -z "$GWP_INSTANCE" ]; then
  GWP_INSTANCE="$GWP_NAME"
fi

kubectl get gatewayproxies.gatewayapi.kubesphere.io -n $GWP_NS $GWP_NAME -o wide
kubectl describe gatewayproxies.gatewayapi.kubesphere.io -n $GWP_NS $GWP_NAME
kubectl get gatewayproxies.gatewayapi.kubesphere.io -n $GWP_NS $GWP_NAME -o yaml
kubectl get pods -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE"
```

GatewayProxy conditions:

| Condition Type         | Status True | Meaning                     |
| ---------------------- | ----------- | --------------------------- |
| `Ready`              | True        | Fully operational           |
| `Progressing`        | True        | Being created or updated    |
| `NewVersionDetected` | True        | New chart version available |

### List Associated Gateways

Each GatewayProxy may have associated standard Gateways and GatewayClasses. The GatewayProxy creates GatewayClasses with the label `gatewayapi.kubesphere.io/gateway-class-name`, and auto-created Gateways carry scope labels:

```bash
# List GatewayClasses created by any GatewayProxy
kubectl get gatewayclass -l "gatewayapi.kubesphere.io/gateway-class-name"

# List Gateways associated with a specific GatewayProxy (by scope label)
kubectl get gateways.gateway.networking.k8s.io -n $GWP_NS \
  -l "gatewayapi.kubesphere.io/scope-type"

# Alternatively, list Gateways by the GatewayClass name they reference
GW_CLASS=$(kubectl get gatewayproxies.gatewayapi.kubesphere.io -n $GWP_NS $GWP_NAME \
  -o jsonpath='{.status.gatewayClass.name}')
kubectl get gateways.gateway.networking.k8s.io -A \
  -o jsonpath='{range .items[?(@.spec.gatewayClassName=="'"$GW_CLASS"'")]}{.metadata.namespace}{"\t"}{.metadata.name}{"\t"}{.spec.gatewayClassName}{"\n"}{end}'
```

### View GatewayProxy Pods and Logs

```bash
kubectl get pods -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE"
kubectl logs -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE" --tail=100
```

---

## Troubleshooting

> Set `$GWP_NAME` and `$GWP_NS` to the target GatewayProxy name and namespace (`kubesphere-controls-system` for all tiers). `$GWP_INSTANCE` is auto-resolved from `status.helmRelease.name` (falls back to `$GWP_NAME`).

### GatewayProxy stuck in `Progressing` state

```bash
kubectl describe gatewayproxies.gatewayapi.kubesphere.io -n $GWP_NS $GWP_NAME

kubectl logs -n extension-gateway-api -l app=gateway-api-controller-manager --tail=200 | grep -iE "(error|helm|install|upgrade|reconcile)"

kubectl get pods -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE"
```

Common causes: Helm chart not found, invalid `rawValues`, Traefik image pull failure, Helm SDK timeout.

### GatewayProxy shows `Ready=False` state

```bash
kubectl describe gatewayproxies.gatewayapi.kubesphere.io -n $GWP_NS $GWP_NAME

kubectl get deployment -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE" -o wide
kubectl describe deployment -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE"
kubectl get pods -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE" -o wide

POD_NAME=$(kubectl get pods -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE" -o jsonpath='{.items[0].metadata.name}')
kubectl describe pod -n $GWP_NS $POD_NAME
kubectl logs -n $GWP_NS $POD_NAME --tail=100
```

Common causes: Image pull failure, resource constraints, port conflicts, missing ConfigMap/Secret, Helm release abnormal.

### GatewayProxy pod crash-looping / CrashLoopBackOff

```bash
kubectl logs -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE" --tail=100 --previous
kubectl get events -n $GWP_NS --sort-by='.lastTimestamp' | tail -20
kubectl exec -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE" -- cat /etc/traefik/traefik.yaml 2>/dev/null | head -50
kubectl get configmap -n $GWP_NS -l "app.kubernetes.io/instance=$GWP_INSTANCE" -o yaml
```

Common causes: Misconfigured Traefik config, port conflicts, resource limits (OOMKilled), missing dependencies (ConfigMap/Secret).

### Controller not reconciling

```bash
kubectl get pods -n extension-gateway-api -l app=gateway-api-controller-manager
kubectl logs -n extension-gateway-api -l app=gateway-api-controller-manager --tail=200
kubectl get validatingwebhookconfiguration -l "app.kubernetes.io/managed-by=Helm,kubesphere.io/extension-ref=gateway-api"
kubectl get deployment -n extension-gateway-api -l app=gateway-api-controller-manager -o yaml
```

Common causes: Controller pod not running, webhook configuration blocking updates, Helm release state mismatch, RBAC permission issues.

### Log search / metrics not working

GatewayProxy observability proxies to `whizard-telemetry-apiserver`:

```bash
kubectl get pods -n extension-whizard-telemetry
kubectl get svc -n extension-whizard-telemetry whizard-telemetry-apiserver
kubectl logs -n extension-gateway-api -l app=gateway-api-apiserver --tail=100 | grep -iE "(log|search|whizard|proxy|metric)"
```

Common causes: Whizard-telemetry not installed or not running, network policy blocking cross-namespace traffic.