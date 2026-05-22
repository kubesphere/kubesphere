---
name: kubesphere-servicemesh
description: KubeSphere ServiceMesh extension management Skill (Istio + Kiali + Jaeger). Use this skill for the ServiceMesh extension Configuration (covers installation, uninstallation, status checks), troubleshooting (covers grayscale release, sidecar injection, topology/metrics, and tracing for Composed Apps Aka Custom Applications).
---

# KubeSphere ServiceMesh

## Overview

Integrates Istio, Kiali, and Jaeger to provide traffic governance for microservices. They operate on Composed App (backed by `applications.app.k8s.io`). A Service is governed when it has the annotation `servicemesh.kubesphere.io/enabled: "true"` and belongs to a Composed App.

Defines two core CRDs:

- **Strategy** — grayscale release task (canary, blue-green, traffic mirroring)
- **ServicePolicy** — traffic management (load balancing, connection pools, circuit breaking, etc.) via Istio DestinationRule

Istio handles traffic routing, Kiali provides topology visualization, and Jaeger enables tracing.

## Before You Start

Check if ServiceMesh is already installed:

```bash
kubectl get installplans.kubesphere.io servicemesh --ignore-not-found
```

If found, upgrading is supported — just select a newer version in Step 1.

## Installation

### Step 1: Detect and Select Version

```bash
ALL_VERSIONS=$(kubectl get extensionversions.kubesphere.io \
  -l kubesphere.io/extension-ref=servicemesh \
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

### Step 3: Check gateway-api Compatibility

```bash
if kubectl get installplans.kubesphere.io gateway-api --ignore-not-found &>/dev/null; then
  GW_API_EXISTS=true
  echo "gateway-api InstallPlan exists → will set PILOT_ENABLE_GATEWAY_API=false"
else
  GW_API_EXISTS=false
  echo "gateway-api not found → no compatibility config needed"
fi

echo "GW_API_EXISTS=$GW_API_EXISTS"
```

If `GW_API_EXISTS=true`, the InstallPlan config will include `PILOT_ENABLE_GATEWAY_API: "false"` to avoid conflicts.

### Step 4: Generate and Apply InstallPlan

```bash
./scripts/generate-installplan.sh "$SELECTED_VERSION" "$TARGET_CLUSTERS" "$GW_API_EXISTS"
```

This generates the YAML to `/tmp/servicemesh-installplan.yaml`, runs `--dry-run=server`, then prints the apply command.

> For configurable extension values (tracing sampling rate, storage backend, credentials, etc.), see [references/extension-values.md](references/extension-values.md).

Apply it:

```bash
kubectl apply -f /tmp/servicemesh-installplan.yaml
```

Tell the user "Installing". Then ask if they want to check status. If yes:

```bash
./scripts/check-status.sh poll
```

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

## Uninstallation

> ⚠ **Always confirm with the user before proceeding.**

### Uninstall from all clusters

```bash
if ! kubectl get installplans.kubesphere.io servicemesh --ignore-not-found &>/dev/null; then
  echo "ServiceMesh is not installed."
  exit 0
fi
```

Confirm with the user, then delete:

```bash
kubectl delete installplans.kubesphere.io servicemesh --ignore-not-found
```

Verify cleanup:

```bash
./scripts/verify-uninstall.sh
```

Success criteria:
1. InstallPlan is deleted
2. No active ServiceMesh pods remain in `extension-servicemesh`

### Uninstall from specific clusters

> **WARNING**: Do NOT delete the InstallPlan. Only remove target clusters from the placement list.

Confirm which clusters to remove, compute remaining clusters, then patch:

```bash
kubectl patch installplans.kubesphere.io servicemesh --type='json' \
  -p='[{"op": "replace", "path": "/spec/clusterScheduling/placement/clusters", "value": ["<REMAINING_CLUSTER_1>", "<REMAINING_CLUSTER_2>"]}]'
```

Success: patch returns OK + removed clusters no longer in `.status.clusterSchedulingStatuses`.

## Troubleshooting

All scenarios below assume the namespace of the Composed App is known. Set it as `NAMESPACE` before proceeding.

### Prerequisites (common steps for all scenarios)

```bash
# 1. Check Composed App status and governance annotation
kubectl -n $NAMESPACE get applications.app.k8s.io -o yaml
#    Key things to verify:
#    - .status: health of composed components
#    - annotation servicemesh.kubesphere.io/enabled=true

# 2. Verify pods with the injection annotation actually have istio-proxy sidecar
kubectl -n $NAMESPACE get pods \
  -l 'app.kubernetes.io/name,app.kubernetes.io/version,app,version' \
  -o custom-columns='NAME:.metadata.name,APPLICATION:.metadata.labels.app\.kubernetes\.io/name,SHOULD_INJECT_ISTIO_PROXY:.metadata.annotations.sidecar\.istio\.io/inject,CONTAINERS:.spec.containers[*].name'

# 3. If sidecar missing, check istiod injection logs
kubectl logs -n extension-servicemesh -l app=istiod --tail=100
```

After completing the prerequisites, proceed to the specific scenario below.

### Grayscale release task (Strategy) not working as expected

When asking the user for `<strategy-name>`, refer to it as "grayscale release task name", not "strategy name".

Data flow: `Strategy → VirtualService → Istio Proxy Sidecar → traffic routing`.

After prerequisites:

```bash
# 1. Check Strategy reconciliation status and events
kubectl -n $NAMESPACE describe strategies.servicemesh.kubesphere.io <strategy-name>

# 2. Check the VirtualService controlled by this Strategy (linked via label)
kubectl -n $NAMESPACE get virtualservice \
  -l "servicemesh.kubesphere.io/controlled-by-strategy=<strategy-name>" -o yaml

# The controller copies strategy.spec.template.spec to the VirtualService spec,
# injecting only route.destination.host, route.destination.port.number, and match.port.
# Compare key fields to verify sync:
#   spec.template.spec.hosts  ↔ spec.hosts
#   spec.template.spec.http   ↔ spec.http  (port/host injected by controller)
#   spec.template.spec.tcp    ↔ spec.tcp   (port/host injected by controller)
# When spec.governor is set, controller overrides all routes to 100% → governor version.
# If the VirtualService spec does not reflect the template, sync failed.

# 3. Verify actual workloads match the routing rules
# (e.g., if routing to version: v2, check pods with that label exist)
kubectl -n $NAMESPACE get pods -l "app=<app-name>,version=<target-version>"

# 4. Check controller-manager sync logs for errors (last 100 lines)
kubectl logs -n extension-servicemesh -l app=servicemesh-controller-manager \
  --tail=100 | grep -iE "(error|reconcile|strategy|<strategy-name>)"
```

### Traffic Monitoring (Topology and metrics) shows no data

Before running commands, confirm with the user:
1. Whether there is actual traffic reaching the Composed App
2. Whether expanding the time range still shows no data

After prerequisites:

```bash
# Check servicemesh-apiserver logs for Kiali access errors
kubectl logs -n extension-servicemesh -l app=servicemesh-apiserver --tail=100
# Check Prometheus / whizard-agent-proxy (only one will exist)
kubectl get pods -n kubesphere-monitoring-system -l 'app.kubernetes.io/name in (prometheus, whizard-agent-proxy)'
```

### Tracing shows no data

Before running commands, confirm with the user:
1. Whether there is actual traffic reaching the Composed App
2. Whether expanding the time range still shows no data

After prerequisites:

```bash
# Check servicemesh-apiserver logs for Jaeger query errors
kubectl logs -n extension-servicemesh -l app=servicemesh-apiserver --tail=100
# Check jaeger-query logs for storage backend connectivity
kubectl logs -n extension-servicemesh -l app.kubernetes.io/component=query --tail=100
# Check jaeger-collector logs for storage backend connectivity
kubectl logs -n extension-servicemesh -l app.kubernetes.io/component=collector --tail=100
# If backend.jaeger.storage.options.es.server-urls is the default, check opensearch
kubectl get pods -n kubesphere-logging-system -l 'app.kubernetes.io/name=opensearch-data'
```