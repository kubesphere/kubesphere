---
name: whizard-auditing
description: Use when working with WizTelemetry Auditing extension for KubeSphere, including installation, configuration, and audit query API
---

# WizTelemetry Auditing

## Overview

WizTelemetry Auditing is an extension component in the KubeSphere Observability Platform for Kubernetes and KubeSphere audit event collection, processing, and storage.

## When to Use

- Installing or configuring the WizTelemetry Auditing extension
- Understanding audit event collection architecture
- Using the audit query API to query audit events

## Components

| Component | Description | Default Enabled |
|-----------|-------------|-----------------|
| kube-auditing | Kubernetes audit event collection and export | true |

## Dependencies

- **WizTelemetry Platform Service** (whizard-telemetry): Required
- **WizTelemetry Data Pipeline** (vector): Required

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
Which clusters do you want to deploy WizTelemetry Auditing to?
```

#### Step 2: Get Latest Version (if not provided by user)

**MUST do this to get the latest version:**

```bash
kubectl get extensionversions -l kubesphere.io/extension-ref=whizard-auditing -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V | tail -1
```

This outputs the latest version (e.g., `1.4.0`). Note this down - you'll use it in the InstallPlan.

### Install WizTelemetry Auditing

**⚠️ IMPORTANT: Complete prerequisite steps BEFORE this step.**

Based on your selections:
- **Target clusters**: User-confirmed cluster names

**⚠️ CRITICAL: InstallPlan `metadata.name` MUST be `whizard-auditing`. DO NOT use any other name.**

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
  name: whizard-auditing
spec:
  extension:
    name: whizard-auditing
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

**Note:** OpenSearch sink configuration (endpoints, auth) is provided by the **vector** extension. Make sure vector is installed and configured with OpenSearch before installing auditing.

#### Enable Doris Sink

To enable Doris sink for audit storage:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-auditing
spec:
  extension:
    name: whizard-auditing
    version: <VERSION>  # From Step 2
  enabled: true
  upgradeStrategy: Manual
  config: |
    kube-auditing:
      sinks:
        opensearch:
          enabled: true
          index:
            prefix: "{{ .cluster }}-auditing"
            timestring: "%Y.%m.%d"
        doris:
          enabled: true
          fe: <DORIS_FE>
          be: <DORIS_BE>
          table:
            partitionUnit: DAY
            retentionPartition: 7
            replicationNum: 2
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

#### Enable ISM Policy

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-auditing
spec:
  extension:
    name: whizard-auditing
    version: <VERSION>  # From Step 2
  enabled: true
  upgradeStrategy: Manual
  config: |
    kube-auditing:
      sinks:
        opensearch:
          enabled: true
          index:
            prefix: "{{ .cluster }}-auditing"
            timestring: "%Y.%m.%d"
      ism_policy:
        enable: true
        min_index_age: "7d"
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

## Configuration Parameters

### OpenSearch Sink Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `kube-auditing.sinks.opensearch.enabled` | bool | true | Enable OpenSearch sink |
| `kube-auditing.sinks.opensearch.index.prefix` | string | "{{ .cluster }}-auditing" | Index prefix |
| `kube-auditing.sinks.opensearch.index.timestring` | string | "%Y.%m.%d" | Index time format |

### Doris Sink Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `kube-auditing.sinks.doris.enabled` | bool | false | Enable Doris sink |
| `kube-auditing.sinks.doris.fe` | string | "" | Doris Frontend address |
| `kube-auditing.sinks.doris.be` | string | "" | Doris Backend address |
| `kube-auditing.sinks.doris.table.partitionUnit` | string | DAY | Partition unit |
| `kube-auditing.sinks.doris.table.retentionPartition` | int | 7 | Retention partition |
| `kube-auditing.sinks.doris.table.replicationNum` | int | 2 | Replication number |

### ISM Policy Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `kube-auditing.ism_policy.enable` | bool | false | Enable Index State Management policy |
| `kube-auditing.ism_policy.min_index_age` | string | "7d" | Minimum index retention period |

## Audit Query API

### Query Audit Events

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.kubesphere.io/v1alpha2/auditing?operation=query&sort=desc&size=10&cluster=host" \
  -H "X-Remote-User: admin"
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `operation` | string | query | Operation type: query/statistics/histogram/export |
| `workspace_filter` | string | | Comma-separated list of workspaces |
| `workspace_search` | string | | Fuzzy match workspace names |
| `objectref_namespace_filter` | string | | Comma-separated list of namespaces (ObjectRef.Namespace) |
| `objectref_namespace_search` | string | | Fuzzy match namespace names |
| `objectref_name_filter` | string | | Comma-separated list of object names |
| `objectref_name_search` | string | | Fuzzy match object names |
| `level_filter` | string | | Audit level: Metadata/Request/RequestResponse |
| `verb_filter` | string | | Comma-separated list of verbs (create, update, delete, etc.) |
| `user_filter` | string | | Comma-separated list of users |
| `user_search` | string | | Fuzzy match username |
| `group_search` | string | | Fuzzy match user groups |
| `source_ip_search` | string | | Fuzzy match source IPs |
| `objectref_resource_filter` | string | | Comma-separated list of resources |
| `objectref_subresource_filter` | string | | Comma-separated list of subresources |
| `response_code_filter` | string | | Comma-separated list of response codes |
| `response_status_filter` | string | | Comma-separated list of response statuses |
| `start_time` | string | | Start time (seconds since epoch) |
| `end_time` | string | | End time (seconds since epoch) |
| `interval` | string | 15m | Time interval for histogram |
| `sort` | string | desc | Sort order: asc/desc |
| `from` | int | 0 | Offset |
| `size` | int | 10 | Number of results |
| `cluster` | string | host | Cluster name |

## Extension Operations

### Check Extension Status

```bash
kubectl get installplan whizard-auditing
kubectl get extensionversions -l kubesphere.io/extension-ref=whizard-auditing
```

### Uninstall Extension

**Uninstall from all clusters:**

```bash
kubectl delete installplan whizard-auditing
```

**Uninstall from specific cluster:**

To remove WizTelemetry Auditing from a specific cluster, update the InstallPlan by removing that cluster from `clusterScheduling.placement.clusters`:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-auditing
spec:
  extension:
    name: whizard-auditing
    version: <VERSION>
  enabled: true
  upgradeStrategy: Manual
  clusterScheduling:
    placement:
      clusters:
        - <REMAINING_CLUSTERS>  # Remove the cluster you want to uninstall from
```
