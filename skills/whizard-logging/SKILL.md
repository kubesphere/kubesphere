---
name: whizard-logging
description: Use when working with WizTelemetry Logging extension for KubeSphere, including installation, configuration, and log query API
---

# WizTelemetry Logging

## Overview

WizTelemetry Logging is an extension component in the KubeSphere Observability Platform for log collection, processing, and storage.

## When to Use

- Installing or configuring the WizTelemetry Logging extension
- Understanding log collection architecture (container logs + disk log collection)
- Using the log query API to query logs

## Components

| Component | Description | Default Enabled |
|-----------|-------------|-----------------|
| vector-logging | Container log collection (collects stdout/stderr from Docker/Containerd) | true |
| logsidecar-injector | Disk log collection (collects logs from files inside containers) | false |

## Dependencies

- **WizTelemetry Platform Service** (whizard-telemetry): Required
- **WizTelemetry Data Pipeline** (vector): Required
- **OpenSearch** (opensearch): Required

## Installation

### Prerequisites

**REQUIRED: Complete all steps in order before generating InstallPlan.**

#### Step 1: Get Available Clusters and Confirm Target

**⚠️ CRITICAL: DO NOT proceed until target clusters are determined.**

**Step 1.1: Get available clusters**

```bash
kubectl get clusters -o jsonpath='{.items[*].metadata.name}'
```

**Step 1.2: Determine target clusters**

- If user **explicitly specified** target clusters in the request → Use those clusters directly, proceed to Step 2
- If user **did NOT specify** target clusters → Ask user to confirm which clusters to deploy to, then proceed to Step 2

**Ask user (if not specified):**
```
Available clusters: host, dev
Which clusters do you want to deploy WizTelemetry Logging to?
```

#### Step 2: Get Latest Version (if not provided by user)

**MUST do this to get the latest version:**

```bash
kubectl get extensionversions -l kubesphere.io/extension-ref=whizard-logging -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V | tail -1
```

This outputs the latest version (e.g., `1.4.0`). Note this down - you'll use it in the InstallPlan.

### Install WizTelemetry Logging

**⚠️ IMPORTANT: Complete prerequisite steps BEFORE this step.**

Based on your selections:
- **Target clusters**: User-confirmed cluster names

**⚠️ CRITICAL: InstallPlan `metadata.name` MUST be `whizard-logging`. DO NOT use any other name.**

**⚠️ CRITICAL: `config` field is YAML format. You MUST:**
- Use the config structure exactly as shown in the template
- **DO NOT** add configuration fields that are not shown in the template
- **DO NOT** modify the structure or hierarchy

**⚠️ CRITICAL: All placeholders MUST be replaced with actual values. DO NOT leave them as placeholders.**

#### Template

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-logging
spec:
  extension:
    name: whizard-logging
    version: <VERSION>  # From Step 2
  enabled: true
  upgradeStrategy: Manual
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

**Replace placeholders:**
- `<VERSION>`: From Step 2 (e.g., `1.4.0`)
- `<TARGET_CLUSTERS>`: User-confirmed cluster names

**Note:** OpenSearch sink configuration (endpoints, auth) is provided by the **vector** extension. Make sure vector is installed and configured with OpenSearch before installing logging.

#### Enable Disk Log Collection

To enable disk log collection, add `logsidecar-injector` to the config:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-logging
spec:
  extension:
    name: whizard-logging
    version: <VERSION>  # From Step 3
  enabled: true
  upgradeStrategy: Manual
  config: |
    logsidecar-injector:
      enabled: true
    vector-logging:
      filter:
        extraLabelSelector: "app.kubernetes.io/name!=kube-events-exporter"
      calico:
        enabled: true
      systemd:
        docker:
          enabled: true
        kubelet:
          enabled: true
      sinks:
        opensearch:
          enabled: true
          index:
            prefix: "{{ .cluster }}-logs"
            timestring: "%Y.%m.%d"
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

## Configuration Parameters

### Logsidecar Injector Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `logsidecar-injector.enabled` | bool | false | Enable disk log collection |
| `logsidecar-injector.sidecar.sidecarType` | string | vector | Sidecar type |
| `logsidecar-injector.resources.limits.cpu` | string | 100m | CPU limit |
| `logsidecar-injector.resources.limits.memory` | string | 100Mi | Memory limit |
| `logsidecar-injector.resources.requests.cpu` | string | 10m | CPU request |
| `logsidecar-injector.resources.requests.memory` | string | 10Mi | Memory request |

### Vector Logging Parameters

#### Filter Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `vector-logging.filter.extraLabelSelector` | string | "app.kubernetes.io/name!=kube-events-exporter" | Extra label selector |
| `vector-logging.filter.extraNamespaceLabelSelector` | string | "" | Extra namespace label selector |
| `vector-logging.filter.includeNamespaces` | list | [] | List of namespaces to collect |
| `vector-logging.filter.excludeNamespaces` | list | [] | List of namespaces to exclude |

#### Calico Log Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `vector-logging.calico.enabled` | bool | true | Enable Calico log collection |
| `vector-logging.calico.logPath` | list | ["/var/log/calico/cni/cni*.log"] | Calico log paths |

#### Systemd Log Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `vector-logging.systemd.docker.enabled` | bool | true | Enable Docker systemd log collection |
| `vector-logging.systemd.kubelet.enabled` | bool | true | Enable Kubelet systemd log collection |
| `vector-logging.systemd.directory` | string | /var/log/journal | Systemd journal directory |

#### OpenSearch Sink Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `vector-logging.sinks.opensearch.enabled` | bool | true | Enable OpenSearch sink |
| `vector-logging.sinks.opensearch.index.prefix` | string | "{{ .cluster }}-logs" | Index prefix |
| `vector-logging.sinks.opensearch.index.timestring` | string | "%Y.%m.%d" | Index time format |

#### ISM Policy Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `vector-logging.ism_policy.enable` | bool | false | Enable Index State Management policy |
| `vector-logging.ism_policy.min_index_age` | string | "7d" | Minimum index retention period |

## Log Query API

### Query Logs

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.kubesphere.io/v1alpha2/logs?operation=query&log_query=error&size=10&cluster=host&sort=desc" \
  -H "X-Remote-User: admin"
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `operation` | string | query | Operation type: query/statistics/histogram/export |
| `namespaces` | string | | Comma-separated list of namespaces |
| `namespace_query` | string | | Fuzzy match namespace names |
| `workloads` | string | | Comma-separated list of workloads |
| `workload_query` | string | | Fuzzy match workload names |
| `pods` | string | | Comma-separated list of pods |
| `pod_query` | string | | Fuzzy match pod names |
| `containers` | string | | Comma-separated list of containers |
| `container_query` | string | | Fuzzy match container names |
| `log_query` | string | | Log content keywords (case-insensitive) |
| `interval` | string | 15m | Time interval for histogram (e.g., 15m, 1h, 1d) |
| `start_time` | string | | Start time (seconds since epoch) |
| `end_time` | string | | End time (seconds since epoch) |
| `sort` | string | desc | Sort order: asc/desc |
| `from` | int | 0 | Offset |
| `size` | int | 10 | Number of results |
| `cluster` | string | host | Cluster name |
| `exportLineLimit` | int | | Max lines for export |

## Extension Operations

### Check Extension Status

```bash
kubectl get installplan whizard-logging
kubectl get extensionversions -l kubesphere.io/extension-ref=whizard-logging
```

### Uninstall Extension

**Uninstall from all clusters:**

```bash
kubectl delete installplan whizard-logging
```

**Uninstall from specific cluster:**

To remove WizTelemetry Logging from a specific cluster, update the InstallPlan by removing that cluster from `clusterScheduling.placement.clusters`:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-logging
spec:
  extension:
    name: whizard-logging
    version: <VERSION>
  enabled: true
  upgradeStrategy: Manual
  clusterScheduling:
    placement:
      clusters:
        - <REMAINING_CLUSTERS>  # Remove the cluster you want to uninstall from
```
