---
name: wiztelemetry-tracing
description: Use when working with WizTelemetry Tracing extension for KubeSphere, including installation, configuration, and tracing query API
---

# WizTelemetry Tracing

## Overview

WizTelemetry Tracing is an extension component in the KubeSphere Observability Platform that provides distributed tracing functionality based on the [OpenTelemetry](https://opentelemetry.io/docs/specs/otel/trace/) standard.

## When to Use

- Installing or configuring the WizTelemetry Tracing extension
- Understanding tracing architecture (Generator + Operator + Collector + Agent)
- Using the tracing query API to query traces, spans, service graphs
- Configuring OpenTelemetry auto-instrumentation for applications

## Components

| Component | Description | Default Enabled |
|-----------|-------------|-----------------|
| generator | WizTelemetry Tracing Generator: generates service graphs from tracing data (StatefulSet) | true |
| operator | OpenTelemetry Operator: manages OpenTelemetry Collector and auto-instrumentation | true |
| collector | OpenTelemetry Collector for WizTelemetry: receives traces, exports to Vector and OpenSearch | true |
| agent | WizTelemetry Tracing Agent: collects tracing data from local log files (DaemonSet) | false |
| demo | OpenTelemetry Demo: generates sample tracing data for demonstration | false |

## Dependencies

- **WizTelemetry Platform Service** (whizard-telemetry): Required
- **OpenSearch** (opensearch): Required

## Installation

### Prerequisites

**REQUIRED: Complete all steps in order before generating InstallPlan.**

#### Step 1: Get Available Clusters and Confirm Target

**CRITICAL: DO NOT proceed until target clusters are determined.**

**Step 1.1: Get available clusters**

```bash
kubectl get clusters -o jsonpath='{.items[*].metadata.name}'
```

**Step 1.2: Determine target clusters**

- If user **explicitly specified** target clusters in the request -> Use those clusters directly, proceed to Step 2
- If user **did NOT specify** target clusters -> Ask user to confirm which clusters to deploy to, then proceed to Step 2

**Ask user (if not specified):**
```
Available clusters: host, dev
Which clusters do you want to deploy WizTelemetry Tracing to?
```

#### Step 2: Get Latest Version (if not provided by user)

**MUST do this to get the latest version:**

```bash
kubectl get extensionversions -n kubesphere-system -l kubesphere.io/extension-ref=wiztelemetry-tracing -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V | tail -1
```

This outputs the latest version (e.g., `1.0.6`). Note this down - you'll use it in the InstallPlan.

#### Step 3: Get Generator Endpoint

**Get a node IP from the target cluster:**

```bash
kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}' 2>/dev/null || \
kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'
```

This outputs a node IP (e.g., `192.168.1.100`). Use this as the generator endpoint.

**If multiple shards (e.g., `generator.shardCount: 3`), generate all endpoints:**

```bash
NODE_IP="<NODE_IP>"
for i in 0 1 2; do
  PORT=$((32318 + i))
  echo "  - http://${NODE_IP}:${PORT}"
done
```

Replace `<NODE_IP>` with the actual node IP from the previous command.

#### Step 4: Confirm Configuration with User

Before creating the InstallPlan, confirm the following with the user:

```
I'll install WizTelemetry Tracing with the following configuration:
- Version: <VERSION>
- Target clusters: <TARGET_CLUSTERS>
- Generator endpoint: http://<NODE_IP>:32318
- OpenSearch endpoint: <OPENSEARCH_ENDPOINT>

Do you want to proceed? (yes/no)
```

If user confirms, proceed to create the InstallPlan. If not, adjust the configuration based on user feedback.

#### Step 5: Create InstallPlan

**CRITICAL: InstallPlan `metadata.name` MUST be `wiztelemetry-tracing`. DO NOT use any other name.**

**CRITICAL: `config` field is YAML format. You MUST:**
- Use the config structure exactly as shown in the template
- **DO NOT** add configuration fields that are not shown in the template
- **DO NOT** modify the structure or hierarchy

**CRITICAL: All placeholders MUST be replaced with actual values. DO NOT leave them as placeholders.**

**Based on your selections:**
- **Target clusters**: User-confirmed cluster names
- **OpenSearch endpoint**: User-provided (default: `https://opensearch-cluster-data.kubesphere-logging-system.svc:9200`)
- **OpenSearch credentials**: User-provided (default user: `admin`)

#### Template

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: wiztelemetry-tracing
  namespace: kubesphere-system
spec:
  extension:
    name: wiztelemetry-tracing
    version: <VERSION>  # From Step 2
  enabled: true
  upgradeStrategy: Manual
  config: |
    global:
      generator:
        endpoints:
          - http://<HOST_NODE_IP>:32318
      storage:
        opensearch:
          auth:
            strategy: basic
            user: <OPENSEARCH_USER>
            password: <OPENSEARCH_PASSWORD>
          endpoints:
            - <OPENSEARCH_ENDPOINT>
    generator:
      shardCount: 1
      service:
        nodePort: 32318
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

**Replace placeholders:**
- `<VERSION>`: From Step 2 (e.g., `1.0.6`)
- `<HOST_NODE_IP>`: Node IP from Step 3 (e.g., `192.168.1.100`)
- `<OPENSEARCH_USER>`: OpenSearch username
- `<OPENSEARCH_PASSWORD>`: OpenSearch password
- `<OPENSEARCH_ENDPOINT>`: OpenSearch endpoint (e.g., `https://opensearch-cluster-data.kubesphere-logging-system.svc:9200`)
- `<TARGET_CLUSTERS>`: User-confirmed cluster names

### Enable Agent (Log File Collection)

To enable tracing data collection from local log files, set `agent.enabled: true` in the config. This deploys a DaemonSet that scans log files and forwards traces to the generator.

### Enable Demo

To enable the OpenTelemetry Demo (generates sample trace data), set `demo.enabled: true` in the config.

### Send Traces to Tempo

To forward traces to Tempo instead of OpenSearch, modify the `collector.collector.config`:

```yaml
collector:
  enabled: true
  collector:
    config: |
      exporters:
        otlp:
          endpoint: <TEMPO_DISTRIBUTOR_GRPC_ENDPOINT>
          headers:
            x-scope-orgid: wiztelemetry-tracing-ks
          tls:
            insecure: true
      service:
        pipelines:
          traces:
            exporters:
              - otlp
              - otlphttp
```

### Multiple Generator Instances

To scale the generator for higher throughput, increase `generator.shardCount` and update `global.generator.endpoints`:

```yaml
global:
  generator:
    endpoints:
      - http://<HOST_NODE_IP>:32318
      - http://<HOST_NODE_IP>:32319
      - http://<HOST_NODE_IP>:32320

generator:
  shardCount: 3
  service:
    nodePort: 32318
```

> `shardCount` must match the number of endpoints in `global.generator.endpoints`.
> The nodePort of each shard increases in sequence starting from the configured `service.nodePort`.

## Configuration Parameters

### Global Parameters

#### DNS Service

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `global.dnsService` | string | `coredns` | DNS service name |

#### Generator Global Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `global.generator.endpoints` | list | `["http://<ip>:32318"]` | Endpoints of all generator shards. Tracing data is routed to different shards by `traceId`. |

#### OpenSearch Storage Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `global.storage.opensearch.auth.strategy` | string | `basic` | Auth strategy |
| `global.storage.opensearch.auth.user` | string | `admin` | OpenSearch username |
| `global.storage.opensearch.auth.password` | string | `admin` | OpenSearch password |
| `global.storage.opensearch.endpoints` | list | | OpenSearch endpoint URLs |
| `global.storage.opensearch.index.prefix` | string | `wiz-tracing-span` | Index prefix |
| `global.storage.opensearch.index.timestring` | string | `%Y.%m.%d` | Index time format (strftime pattern) |

### Generator Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `generator.shardCount` | int | `1` | Number of generator shards. Must equal the number of entries in `global.generator.endpoints`. |
| `generator.image.tag` | string | `v1.0.2` | Generator image tag |
| `generator.image.imagePullPolicy` | string | `IfNotPresent` | Image pull policy |
| `generator.service.nodePort` | int | `32318` | NodePort of the first shard. Subsequent shards use consecutive ports. |
| `generator.resources.limits.cpu` | string | `2` | CPU limit |
| `generator.resources.limits.memory` | string | `2000Mi` | Memory limit |
| `generator.resources.requests.cpu` | string | `100m` | CPU request |
| `generator.resources.requests.memory` | string | `100Mi` | Memory request |

### Operator Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `operator.enabled` | bool | `true` | Enable OpenTelemetry Operator |
| `operator.fullnameOverride` | string | `wiztelemetry-tracing-operator` | Override name for the operator deployment |

### Collector Parameters

#### ISM Policy Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `collector.ism_policy.enable` | bool | `true` | Enable OpenSearch Index State Management policy |
| `collector.ism_policy.span_index_pattern` | string | `*wiz-tracing-span*` | Index pattern for span indices |
| `collector.ism_policy.service_index_pattern` | string | `*wiz-tracing-service*` | Index pattern for service indices |
| `collector.ism_policy.min_index_age` | string | `7d` | Minimum index retention period |
| `collector.ism_policy.span_index_priority` | int | `9950` | Index priority for span indices |

#### Collector Deployment Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `collector.enabled` | bool | `true` | Enable OpenTelemetry Collector |
| `collector.collector.mode` | string | `deployment` | Deployment mode (`deployment` or `daemonset`) |
| `collector.collector.replicas` | int | `2` | Number of collector replicas |
| `collector.collector.upgradeStrategy` | string | `automatic` | Upgrade strategy |

### Agent Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `agent.enabled` | bool | `false` | Enable Tracing Agent (DaemonSet) |
| `agent.dockerRootDir` | string | `/var/lib/docker` | Docker root directory |
| `agent.config.scanInterval` | string | `1m` | Log file scan interval |
| `agent.config.deletionDelay` | string | `5m` | Delay before deleting processed files |
| `agent.config.logPath` | string | `/app/logs/*/*trace*.log` | Path pattern for trace log files |
| `agent.config.podLabelSelector` | map | `{}` | Filter pods by label selector |
| `agent.config.namespaceLabelSelector` | map | `{}` | Filter namespaces by label selector |
| `agent.config.includeNamespaces` | list | `[]` | List of namespaces to include |
| `agent.config.excludeNamespaces` | list | `[]` | List of namespaces to exclude |
| `agent.resources.limits.cpu` | string | `500m` | CPU limit |
| `agent.resources.limits.memory` | string | `500Mi` | Memory limit |
| `agent.resources.requests.cpu` | string | `10m` | CPU request |
| `agent.resources.requests.memory` | string | `10Mi` | Memory request |
| `agent.vector.resources.limits.cpu` | string | `2` | Vector sidecar CPU limit |
| `agent.vector.resources.limits.memory` | string | `2000Mi` | Vector sidecar memory limit |
| `agent.vector.resources.requests.cpu` | string | `100m` | Vector sidecar CPU request |
| `agent.vector.resources.requests.memory` | string | `100Mi` | Vector sidecar memory request |

### Demo Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `demo.enabled` | bool | `false` | Enable OpenTelemetry Demo |

## Tracing Query API

### Query Traces

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/tracing.wiztelemetry.io/v1alpha1/traces" \
  -H "X-Remote-User: admin" \
  -H "Content-Type: application/json" \
  -d '{"condition": {}, "limit": 20}'
```

### Query Spans

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/tracing.wiztelemetry.io/v1alpha1/spans" \
  -H "X-Remote-User: admin" \
  -H "Content-Type: application/json" \
  -d '{"condition": {}, "limit": 20}'
```

### Get Service Graph

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/tracing.wiztelemetry.io/v1alpha1/servicegraphs" \
  -H "X-Remote-User: admin" \
  -H "Content-Type: application/json" \
  -d '{"condition": {"start": "<START_TIME>", "end": "<END_TIME>"}}'
```

### Get Services

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/tracing.wiztelemetry.io/v1alpha1/services" \
  -H "X-Remote-User: admin"
```

### Get Tags

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/tracing.wiztelemetry.io/v1alpha1/tags" \
  -H "X-Remote-User: admin"
```

### Get Values by Tag

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/tracing.wiztelemetry.io/v1alpha1/values?tags=span.kind&limit=100" \
  -H "X-Remote-User: admin"
```

### Get Histogram

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/tracing.wiztelemetry.io/v1alpha1/histogram?key=<KEY>&kind=1&interval=15m&startTime=<START>&endTime=<END>" \
  -H "X-Remote-User: admin"
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `key` | string | The key of node or edge |
| `kind` | int | Histogram kind: 1=RequestTotal, 2=RequestTimeAverage, 3=FailedRequestTotal, 4=ResponseTotal, 5=ResponseTimeAverage, 6=FailedResponseTotal |
| `interval` | string | Time interval (e.g., `15m`, `1h`, `1d`) |
| `startTime` | string | Start time in seconds since epoch |
| `endTime` | string | End time in seconds since epoch |

### Get Associated Workloads

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/tracing.wiztelemetry.io/v1alpha1/workloads?keys=<KEYS>&startTime=<START>&endTime=<END>" \
  -H "X-Remote-User: admin"
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `keys` | string | Comma-separated list of node keys |
| `startTime` | string | Start time in seconds since epoch |
| `endTime` | string | End time in seconds since epoch |

### Get/Set Apdex Thresholds

```bash
# Get apdex thresholds
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/tracing.wiztelemetry.io/v1alpha1/apdex/thresholds?keys=<KEYS>" \
  -H "X-Remote-User: admin"

# Set apdex thresholds
curl -X PUT "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/tracing.wiztelemetry.io/v1alpha1/apdex/thresholds" \
  -H "X-Remote-User: admin" \
  -H "Content-Type: application/json" \
  -d '{"serviceA:serviceB": 0.5}'
```

## OpenTelemetry Auto-Instrumentation

### Create Instrumentation Resource

```bash
kubectl apply -f - <<EOF
apiVersion: opentelemetry.io/v1alpha1
kind: Instrumentation
metadata:
  name: my-instrumentation
spec:
  exporter:
    endpoint: http://wiztelemetry-tracing-collector.wiz-telemetry-tracing:4317
  propagators:
    - tracecontext
    - baggage
    - b3
  sampler:
    type: parentbased_traceidratio
    argument: "0.25"
  python:
    env:
      - name: OTEL_EXPORTER_OTLP_ENDPOINT
        value: http://wiztelemetry-tracing-collector.wiz-telemetry-tracing:4318
  dotnet:
    env:
      - name: OTEL_EXPORTER_OTLP_ENDPOINT
        value: http://wiztelemetry-tracing-collector.wiz-telemetry-tracing:4318
  go:
    env:
      - name: OTEL_EXPORTER_OTLP_ENDPOINT
        value: http://wiztelemetry-tracing-collector.wiz-telemetry-tracing:4318
EOF
```

### Enable Instrumentation via Annotations

Add the following annotation to a pod or namespace:

| Language | Annotation |
|----------|-----------|
| Java | `instrumentation.opentelemetry.io/inject-java: "true"` |
| NodeJS | `instrumentation.opentelemetry.io/inject-nodejs: "true"` |
| Python | `instrumentation.opentelemetry.io/inject-python: "true"` |
| .NET | `instrumentation.opentelemetry.io/inject-dotnet: "true"` |
| Go | `instrumentation.opentelemetry.io/inject-go: "true"` |
| Apache HTTPD | `instrumentation.opentelemetry.io/inject-apache-httpd: "true"` |
| Nginx | `instrumentation.opentelemetry.io/inject-nginx: "true"` |
| SDK only | `instrumentation.opentelemetry.io/inject-sdk: "true"` |

Annotation values:
- `"true"` - inject from the namespace's Instrumentation resource
- `"my-instrumentation"` - use a specific Instrumentation CR in the current namespace
- `"my-ns/my-instrumentation"` - use an Instrumentation CR in a different namespace
- `"false"` - do not inject

## Extension Operations

### Check Extension Status

```bash
kubectl get installplan wiztelemetry-tracing
kubectl get extensions wiztelemetry-tracing
```

### Check Component Status

```bash
kubectl get statefulset -n wiz-telemetry-tracing
kubectl get deployment -n wiz-telemetry-tracing
kubectl get daemonset -n wiz-telemetry-tracing
kubectl get pods -n wiz-telemetry-tracing
```

### Uninstall Extension

**Uninstall from all clusters:**

```bash
kubectl delete installplan wiztelemetry-tracing
```

**Uninstall from specific cluster:**

To remove WizTelemetry Tracing from a specific cluster, update the InstallPlan by removing that cluster from `clusterScheduling.placement.clusters`:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: wiztelemetry-tracing
spec:
  extension:
    name: wiztelemetry-tracing
    version: <VERSION>
  enabled: true
  upgradeStrategy: Manual
  clusterScheduling:
    placement:
      clusters:
        - <REMAINING_CLUSTERS>  # Remove the cluster you want to uninstall from
```
