---
name: whizard-events
description: Use when working with WizTelemetry Events extension for KubeSphere, including installation, configuration, and event query API
---

# WizTelemetry Events

## Overview

WizTelemetry Events is an extension component in the KubeSphere Observability Platform for Kubernetes event collection, processing, and storage.

## When to Use

- Installing or configuring the WizTelemetry Events extension
- Understanding event collection architecture
- Using the event query API to query events

## Components

| Component | Description | Default Enabled |
|-----------|-------------|-----------------|
| kube-events-exporter | Kubernetes event collection and export | true |

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
Which clusters do you want to deploy WizTelemetry Events to?
```

#### Step 2: Get Latest Version (if not provided by user)

**MUST do this to get the latest version:**

```bash
kubectl get extensionversions -l kubesphere.io/extension-ref=whizard-events -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V | tail -1
```

This outputs the latest version (e.g., `1.4.0`). Note this down - you'll use it in the InstallPlan.

### Install WizTelemetry Events

**⚠️ IMPORTANT: Complete prerequisite steps BEFORE this step.**

Based on your selections:
- **Target clusters**: User-confirmed cluster names

**⚠️ CRITICAL: InstallPlan `metadata.name` MUST be `whizard-events`. DO NOT use any other name.**

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
  name: whizard-events
spec:
  extension:
    name: whizard-events
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

**Note:** OpenSearch sink configuration (endpoints, auth) is provided by the **vector** extension. Make sure vector is installed and configured with OpenSearch before installing events.

#### Enable ISM Policy

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-events
  namespace: kubesphere-system
spec:
  extension:
    name: whizard-events
    version: <VERSION>  # From Step 2
  enabled: true
  upgradeStrategy: Manual
  config: |
    kube-events-exporter:
      sinks:
        opensearch:
          enabled: true
          index:
            prefix: "{{ .cluster }}-events"
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
| `kube-events-exporter.sinks.opensearch.enabled` | bool | true | Enable OpenSearch sink |
| `kube-events-exporter.sinks.opensearch.index.prefix` | string | "{{ .cluster }}-events" | Index prefix |
| `kube-events-exporter.sinks.opensearch.index.timestring` | string | "%Y.%m.%d" | Index time format |

### ISM Policy Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `kube-events-exporter.ism_policy.enable` | bool | false | Enable Index State Management policy |
| `kube-events-exporter.ism_policy.min_index_age` | string | "7d" | Minimum index retention period |

## Event Query API

### Query Events

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.kubesphere.io/v1alpha2/events?operation=query&sort=desc&size=10&cluster=host" \
  -H "X-Remote-User: admin"
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `operation` | string | query | Operation type: query/statistics/histogram/export |
| `workspace_filter` | string | | Comma-separated list of workspaces |
| `workspace_search` | string | | Fuzzy match workspace names |
| `involved_object_namespace_filter` | string | | Comma-separated list of namespaces (involvedObject.namespace) |
| `involved_object_namespace_search` | string | | Fuzzy match namespace names |
| `involved_object_name_filter` | string | | Comma-separated list of object names |
| `involved_object_name_search` | string | | Fuzzy match object names |
| `involved_object_kind_filter` | string | | Comma-separated list of kinds |
| `reason_filter` | string | | Comma-separated list of reasons |
| `reason_search` | string | | Fuzzy match reason |
| `message_search` | string | | Fuzzy match message |
| `type_filter` | string | | Event type: Warning/Normal |
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
kubectl get installplan whizard-events
kubectl get extensionversions -l kubesphere.io/extension-ref=whizard-events
```

### Uninstall Extension

**Uninstall from all clusters:**

```bash
kubectl delete installplan whizard-events
```

**Uninstall from specific cluster:**

To remove WizTelemetry Events from a specific cluster, update the InstallPlan by removing that cluster from `clusterScheduling.placement.clusters`:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-events
spec:
  extension:
    name: whizard-events
    version: <VERSION>
  enabled: true
  upgradeStrategy: Manual
  clusterScheduling:
    placement:
      clusters:
        - <REMAINING_CLUSTERS>  # Remove the cluster you want to uninstall from
```
