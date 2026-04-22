---
name: vector
description: Use when installing or configuring the WizTelemetry Data Pipeline (vector) extension for KubeSphere, which provides data collection, transformation, and routing for observability data including logs, auditing, events, and notifications
---

# WizTelemetry Data Pipeline (Vector)

## Overview

WizTelemetry Data Pipeline is an extension based on vector (https://vector.dev/) that provides the ability to collect, transform, and route observability data. It is a core dependency for other WizTelemetry extensions like Logging, Auditing, Events, and Notification.

## When to Use

- Installing or configuring the WizTelemetry Data Pipeline extension
- Setting up data collection for logs, auditing, events, and notifications
- Configuring Vector sinks (OpenSearch)
- Managing Vector agent components

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
Which clusters do you want to deploy Vector to?
```

#### Step 2: Get OpenSearch Endpoint and Credentials (MUST DO)

- If user **already provided** OpenSearch endpoint and credentials in the request → Use those directly, proceed to Step 3
- If user **did NOT provide** → **You MUST ask user** for OpenSearch endpoint and credentials

**Ask user for (if not provided):**

1. **OpenSearch endpoint URL** (required)
   - Example: `http://<node-ip>:30920` or `https://opensearch.example.com:9200`

2. **OpenSearch credentials** (required)
   - Username (default: `admin`)
   - Password

**DO NOT proceed to Step 3 until user provides both endpoint and credentials.**

#### Step 3: Get Latest Vector Version (if not provided by user)

**MUST do this to get the latest version:**

```bash
kubectl get extensionversions -l kubesphere.io/extension-ref=vector -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V | tail -1
```

This outputs the latest version (e.g., `1.1.4`). Note this down - you'll use it in the InstallPlan.

### Install Vector Extension

**⚠️ IMPORTANT: Complete prerequisite steps (1-3) BEFORE this step.**

**⚠️ CRITICAL: InstallPlan `metadata.name` MUST be `vector`. DO NOT use any other name.**

Based on your selections:
- **Target clusters**: Use the user-confirmed cluster names
- **OpenSearch endpoint**: User-provided endpoint
- **OpenSearch credentials**: User-provided username and password

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
  name: vector
spec:
  extension:
    name: vector
    version: <VECTOR_VERSION>  # From Step 3
  enabled: true
  upgradeStrategy: Manual
  config: |
    agent:
      sinks:
        opensearch:
          auth:
            strategy: basic
            user: <OPENSEARCH_USER>
            password: <OPENSEARCH_PASSWORD>
          endpoints:
            - <OPENSEARCH_ENDPOINT>
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

**Replace placeholders:**
- `<VECTOR_VERSION>`: From Step 2 (e.g., `1.1.4`)
- `<OPENSEARCH_ENDPOINT>`: User-provided endpoint (e.g., `http://<node-ip>:30920`)
- `<OPENSEARCH_USER>`: User-provided username (default: `admin`)
- `<OPENSEARCH_PASSWORD>`: User-provided password
- `<TARGET_CLUSTERS>`: User-confirmed cluster names

**⚠️ DO NOT generate InstallPlan until all placeholders have real values.**

### Wait for Deployment

**After applying InstallPlan, you MUST wait for deployment to complete:**

```bash
# Wait for Vector pods to be ready (on each cluster)
kubectl wait --for=condition=Ready pods -n kubesphere-logging-system -l app.kubernetes.io/instance=vector --timeout=300s

# Verify deployment status
kubectl get pods -n kubesphere-logging-system -l app.kubernetes.io/instance=vector
```

**Show deployment summary to user:**
- Which clusters Vector was deployed to
- OpenSearch endpoint used
- Pod status (Ready/Total)

#### Enable Metrics Export

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: vector
spec:
  extension:
    name: vector
    version: <VECTOR_VERSION>  # From Step 2
  enabled: true
  upgradeStrategy: Manual
  config: |
    agent:
      sinks:
        opensearch:
          auth:
            strategy: basic
            user: <OPENSEARCH_USER>
            password: <OPENSEARCH_PASSWORD>
          endpoints:
            - <OPENSEARCH_ENDPOINT>
      exportMetrics:
        enabled: true
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

## Configuration Parameters

### Agent Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `agent.role` | string | "Agent" | Role identifier |
| `agent.image.tag` | string | "0.53.0-debian" | Vector image tag |
| `agent.resources.requests.cpu` | string | "100m" | CPU request |
| `agent.resources.requests.memory` | string | "100Mi" | Memory request |
| `agent.resources.limits.cpu` | string | "2000m" | CPU limit |
| `agent.resources.limits.memory` | string | "2000Mi" | Memory limit |
| `agent.service.ports` | list | see values.yaml | Service ports |
| `agent.exportMetrics.enabled` | bool | false | Enable metrics export |

### Agent Sinks OpenSearch Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `agent.sinks.opensearch.endpoints` | list | Yes | OpenSearch endpoint URLs |
| `agent.sinks.opensearch.auth.strategy` | string | Yes | Authentication strategy (set to `basic`) |
| `agent.sinks.opensearch.auth.user` | string | Yes | Username for authentication |
| `agent.sinks.opensearch.auth.password` | string | Yes | Password for authentication |
| `agent.sinks.opensearch.tls.verify` | bool | No | Enable TLS verification (default: false) |

**Example:**

```yaml
agent:
  sinks:
    opensearch:
      endpoints:
        - http://<node-ip>:30920
      auth:
        strategy: basic
        user: admin
        password: admin
      tls:
        verify: false
```

### Docker Root Directory Configuration

If Docker root directory is not `/var/lib`:

```yaml
agent:
  extraVolumes:
    - name: docker-root
      hostPath:
        path: /path/to/docker
        type: ''
  extraVolumeMounts:
    - name: docker-root
      mountPath: /path/to/docker
```

## Extension Operations

### Check Extension Status

```bash
# View extension installation status
kubectl get installplan vector

# View extension version
kubectl get extensionversions -l kubesphere.io/extension-ref=vector
```

### Check Pod Status

```bash
# View all Vector pods
kubectl get pods -n kubesphere-logging-system -l app.kubernetes.io/name=vector

# View agent pods
kubectl get pods -n kubesphere-logging-system -l app.kubernetes.io/name=vector,app.kubernetes.io/component=agent
```

### View Logs

```bash
# View agent logs
kubectl logs -n kubesphere-logging-system -l app.kubernetes.io/name=vector,app.kubernetes.io/component=agent --tail=100
```

### Update Configuration

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: vector
spec:
  extension:
    name: vector
    version: <VECTOR_VERSION>
  enabled: true
  upgradeStrategy: Manual
  config: |
    agent:
      sinks:
        opensearch:
          auth:
            strategy: basic
            user: <OPENSEARCH_USER>
            password: <OPENSEARCH_PASSWORD>
          endpoints:
            - <OPENSEARCH_ENDPOINT>
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

### Uninstall Extension

**Uninstall from all clusters:**

```bash
kubectl delete installplan vector
```

**Uninstall from specific cluster:**

To remove Vector from a specific cluster, update the InstallPlan by removing that cluster from `clusterScheduling.placement.clusters`:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: vector
spec:
  extension:
    name: vector
    version: <VECTOR_VERSION>
  enabled: true
  upgradeStrategy: Manual
  config: |
    agent:
      sinks:
        opensearch:
          auth:
            strategy: basic
            user: <OPENSEARCH_USER>
            password: <OPENSEARCH_PASSWORD>
          endpoints:
            - <OPENSEARCH_ENDPOINT>
  clusterScheduling:
    placement:
      clusters:
        - <REMAINING_CLUSTERS>  # Remove the cluster you want to uninstall from
```

## Important Notes

1. **Dependency**: Vector is a core dependency for WizTelemetry extensions. Install it first before installing Logging, Auditing, Events, or Notification.
2. **OpenSearch Required**: User must provide OpenSearch endpoint and credentials.
3. **Multicluster**: The extension uses `installationMode: Multicluster`:
   - `agent` (tag: agent) is deployed to all selected member clusters
4. **Agent Scheduling**: Agent pods have affinity to avoid edge nodes and tolerate all taints.
5. **Cross-cluster Access**: Ensure OpenSearch endpoint is accessible from all Vector clusters.

## Troubleshooting

### Check Vector Configuration

```bash
# View Vector configmap
kubectl get configmap -n kubesphere-logging-system -l app.kubernetes.io/name=vector

# View specific config
kubectl get configmap -n kubesphere-logging-system vector-config -o yaml
```

### Verify Sinks

```bash
# Check if sinks are configured correctly
kubectl get secret -n kubesphere-logging-system vector-sinks -o yaml
```

### Common Issues

| Issue | Solution |
|-------|----------|
| Pods not starting | Check if OpenSearch is accessible |
| Data not flowing | Verify sink configuration and network connectivity |
| Agent not on member cluster | Check multicluster installation settings |
| Out of memory | Increase resource limits in configuration |
