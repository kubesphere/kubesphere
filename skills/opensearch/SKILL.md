---
name: opensearch
description: Use when installing or configuring the OpenSearch extension for KubeSphere, which provides distributed search and analytics engine for storing logs, events, auditing, and notification history
---

# OpenSearch Extension

## Overview

OpenSearch is a distributed search and analytics engine built into the KubeSphere WizTelemetry Observability Platform. It is used to store, search, and analyze observability data including logs, auditing events, K8s events, and notification history.

## When to Use

- Installing OpenSearch cluster for KubeSphere observability extensions
- Configuring OpenSearch storage for logging, events, auditing, and notifications
- Managing OpenSearch Dashboard and Curator components

## Components

| Component | Description | Default |
|-----------|-------------|---------|
| opensearch-master | Master node for cluster coordination | 1 replica |
| opensearch-data | Data nodes for storing indices | 3 replicas |
| opensearch-dashboards | Web UI for visualizing data | disabled |
| opensearch-curator | Scheduled task to clean old indices | enabled |

## Prerequisites

### Check Installation Status

```bash
# Check if OpenSearch is installed
kubectl get installplan opensearch -o jsonpath='{.spec.enabled}'

# Get installed version
kubectl get extension opensearch -o jsonpath='{.status.installedVersion}'

# Get target clusters
kubectl get installplan opensearch -o jsonpath='{.spec.clusterScheduling.placement.clusters}'
```

Returns:
- `"true"` - installed and enabled
- `"false"` - installed but disabled
- Empty/Error - not installed

### Get Available Clusters

```bash
kubectl get clusters -o jsonpath='{.items[*].metadata.name}'
```

### Confirm Target Clusters (MUST DO)

**⚠️ CRITICAL: Do Not guess.**

**⚠️ CRITICAL: DO NOT proceed until target clusters are determined.**

- If user **explicitly specified** target clusters in the request → Use those clusters directly
- If user **did NOT specify** target clusters → You MUST ask user to confirm which clusters to deploy to

**Ask user (if not specified):**
```
Available clusters: host, dev
Which clusters do you want to deploy OpenSearch to?
```

### Get Latest Version

**MUST do this to get the latest version:**

```bash
kubectl get extensionversions -n kubesphere-system -l kubesphere.io/extension-ref=opensearch -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V | tail -1
```

## Installation

**⚠️ CRITICAL: InstallPlan `metadata.name` MUST be `opensearch`. DO NOT use any other name.**

Use the latest version obtained from "Get Latest Version" step.

### Minimal Installation

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: opensearch
spec:
  extension:
    name: opensearch
    version: <VERSION>  # From Get Latest Version step
  enabled: true
  upgradeStrategy: Manual
  clusterScheduling:
    placement:
      clusters:
        - host
```

### With OpenSearch Dashboard

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: opensearch
spec:
  extension:
    name: opensearch
    version: <VERSION>  # From Get Latest Version step
  enabled: true
  upgradeStrategy: Manual
  config: |
    opensearch-dashboards:
      enabled: true
  clusterScheduling:
    placement:
      clusters:
        - host
```

### Disable Curator

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: opensearch
spec:
  extension:
    name: opensearch
    version: <VERSION>  # From Get Latest Version step
  enabled: true
  upgradeStrategy: Manual
  config: |
    opensearch-curator:
      enabled: false
  clusterScheduling:
    placement:
      clusters:
        - host
```

## Configuration

### Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `opensearch-master.replicas` | int | 1 | Number of master nodes |
| `opensearch-master.opensearchJavaOpts` | string | "-Xmx512M -Xms512M" | JVM options |
| `opensearch-master.resources` | object | - | Resource limits/requests |
| `opensearch-data.replicas` | int | 3 | Number of data nodes |
| `opensearch-data.opensearchJavaOpts` | string | "-Xmx1536M -Xms1536M" | JVM options |
| `opensearch-data.resources` | object | - | Resource limits/requests |
| `opensearch-data.service.type` | string | NodePort | Service type |
| `opensearch-data.service.nodePort` | int | 30920 | NodePort for external access |
| `opensearch-dashboards.enabled` | bool | false | Enable OpenSearch Dashboards |
| `opensearch-curator.enabled` | bool | true | Enable index cleanup job |

### Resource Configuration

```yaml
opensearch-master:
  replicas: 1
  opensearchJavaOpts: "-Xmx512M -Xms512M"
  resources:
    requests:
      cpu: "100m"
      memory: "512Mi"
    limits:
      cpu: "500m"
      memory: "1Gi"

opensearch-data:
  replicas: 3
  opensearchJavaOpts: "-Xmx2048M -Xms2048M"
  resources:
    requests:
      cpu: "200m"
      memory: "2Gi"
    limits:
      cpu: "2000m"
      memory: "4Gi"
  persistence:
    size: 50Gi
    storageClass: "local-volume"
```

## Operations

### Access Target Cluster

> ⚠️ **CRITICAL**: You MUST access the target cluster BEFORE checking status or running any kubectl commands below.
> 
> ❌ **DO NOT** run kubectl commands without using the target cluster kubeconfig
> ✅ **MUST** run: `kubectl --kubeconfig=/tmp/<cluster>-kubeconfig ...`

**Checklist (must complete before any operation):**

- [ ] 1. Find target clusters:
  ```bash
  kubectl get installplan opensearch -o jsonpath='{.spec.clusterScheduling.placement.clusters}'
  ```

- [ ] 2. Get cluster kubeconfig:
  ```bash
  kubectl get cluster <cluster-name> -o jsonpath='{.spec.connection.kubeconfig}' | base64 -d > /tmp/<cluster-name>-kubeconfig
  ```

- [ ] 3. Use this kubeconfig for ALL subsequent commands

> **Note**: Replace `<cluster-name>` with the actual cluster name. Use `spec.connection.kubeconfig` for imported clusters.

### Check Status

> ⚠️ **REQUIRED**: Complete [Access Target Cluster](#access-target-cluster) checklist FIRST

```bash
# Extension status
kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get extension opensearch

# InstallPlan status
kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get installplan opensearch
```

### Get Pods

> ⚠️ **REQUIRED**: Complete [Access Target Cluster](#access-target-cluster) checklist FIRST

```bash
# All pods
kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get pods -n kubesphere-logging-system -l app.kubernetes.io/instance=opensearch-agent

# By component
kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get pods -n kubesphere-logging-system -l app.kubernetes.io/name=opensearch-master
kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get pods -n kubesphere-logging-system -l app.kubernetes.io/name=opensearch-data

# Status
kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get pods -n kubesphere-logging-system -l app.kubernetes.io/instance=opensearch-agent -o wide
```

### External Endpoint

> ⚠️ **REQUIRED**: Complete [Access Target Cluster](#access-target-cluster) checklist FIRST

OpenSearch uses NodePort (default: 30920) to expose service externally.

```bash
# Get NodePort service
kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get svc -n kubesphere-logging-system opensearch-cluster-data

# Get node IP
kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'

# Access OpenSearch externally
# URL: https://<node-ip>:30920
curl -k -u admin:admin "https://<node-ip>:30920/_cluster/health"
```

---

### Output Contract (For Other Skills)

> ⚠️ **This section defines what information this skill provides to other skills**

When OpenSearch is successfully deployed, other skills can retrieve:

| Field | Example Value | Description |
|-------|---------------|-------------|
| `endpoint` | `https://127.0.0.1:30920` | Full URL for Vector sink |
| `nodePort` | `30920` | NodePort number |
| `nodeIP` | `127.0.0.1` | Node internal IP |
| `auth.user` | `admin` | Default username |
| `auth.password` | `admin` | Default password |

**How other skills should get this info:**

1. Find which cluster OpenSearch was deployed to:
   ```bash
   kubectl get installplan opensearch -o jsonpath='{.spec.clusterScheduling.placement.clusters}'
   ```

2. Get cluster kubeconfig:
   ```bash
   kubectl get cluster <cluster-name> -o jsonpath='{.spec.connection.kubeconfig}' | base64 -d > /tmp/<cluster-name>-kubeconfig
   ```

3. Get OpenSearch endpoint:
   ```bash
   # NodePort
   NODE_PORT=$(kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get svc -n kubesphere-logging-system opensearch-cluster-data -o jsonpath='{.spec.ports[?(@.port==9200)].nodePort}')
   
   # Node IP
   NODE_IP=$(kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
   
   # Full endpoint
   echo "https://${NODE_IP}:${NODE_PORT}"
   ```

### Update Configuration

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: opensearch
spec:
  extension:
    name: opensearch
    version: <VERSION>  # From Get Latest Version step
  enabled: true
  upgradeStrategy: Manual
  config: |
    opensearch-data:
      replicas: 5
      opensearchJavaOpts: "-Xmx4096M -Xms4096M"
    opensearch-dashboards:
      enabled: true
  clusterScheduling:
    placement:
      clusters:
        - host
```

### Uninstall

**Uninstall from all clusters:**

```bash
kubectl delete installplan opensearch
```

**Uninstall from specific cluster:**

To remove OpenSearch from a specific cluster, update the InstallPlan by removing that cluster from `clusterScheduling.placement.clusters`:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: opensearch
spec:
  extension:
    name: opensearch
    version: <VERSION>
  enabled: true
  upgradeStrategy: Manual
  config: |
    opensearch-data:
      replicas: 3
  clusterScheduling:
    placement:
      clusters:
        - <REMAINING_CLUSTERS>  # Remove the cluster you want to uninstall from
```

## Troubleshooting

> ⚠️ **REQUIRED**: Complete [Access Target Cluster](#access-target-cluster) checklist FIRST

### Pod Issues

```bash
kubectl --kubeconfig=/tmp/<cluster>-kubeconfig get pods -n kubesphere-logging-system -l app.kubernetes.io/instance=opensearch-agent

# Pod events
kubectl --kubeconfig=/tmp/<cluster>-kubeconfig describe pods -n kubesphere-logging-system -l app.kubernetes.io/instance=opensearch-agent
```

### OpenSearch Health

> ⚠️ **REQUIRED**: Complete [Access Target Cluster](#access-target-cluster) checklist FIRST

```bash
# Cluster health
curl -k -u admin:admin "https://<node-ip>:30920/_cluster/health"

# List indices
curl -k -u admin:admin "https://<node-ip>:30920/_cat/indices"
```

### Common Issues

| Issue | Solution |
|-------|----------|
| Pods not starting | Check node resources and storage availability |
| Out of memory | Increase JVM heap size |
| Cannot connect | Check NodePort firewall rules |
| Index storage full | Increase PVC size or enable Curator |

## Notes

1. **Resource Requirements**: OpenSearch requires significant CPU and memory
2. **Storage**: Data nodes use persistent storage
3. **NodePort**: Default port is 30920
4. **Multi-cluster**: Can be deployed to specific clusters
