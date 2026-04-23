---
name: whizard-telemetry-ruler
description: Use when working with WizTelemetry Ruler extension for KubeSphere, including installation, configuration, alerting rules management
---

# WizTelemetry Ruler

## Overview

WizTelemetry Ruler is an extension component in the KubeSphere Observability Platform that provides event alerting and log alerting capabilities. It can define alerting rules for K8s native events, K8s/KubeSphere auditing events, and K8s logs, evaluate incoming event data and log data, and send alerts to specified receivers such as alertmanager, etc.

## When to Use

- Installing or configuring the WizTelemetry Ruler extension
- Creating, updating, or deleting alerting rules (RuleGroup/ClusterRuleGroup)
- Managing alerting configurations
- Using the ruler API to manage alerting rules

## Components

| Component | Description | Default Enabled |
|-----------|-------------|-----------------|
| whizard-telemetry-ruler | Core ruler component for alerting | true |

## Dependencies

- **WizTelemetry Platform Service** (whizard-telemetry): Required
- **WizTelemetry Events** (whizard-events): Required if event alerting is enabled
- **WizTelemetry Auditing** (whizard-auditing): Required if auditing alerting is enabled
- **WizTelemetry Logging** (whizard-logging): Required if logging alerting is enabled
- **WizTelemetry Notification** (whizard-notification): Optional (for alert notification)
- **WizTelemetry Data Pipeline** (vector): Required if alerting persistence is enabled
- **OpenSearch** (opensearch): Required if alerting persistence is enabled
- 
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
Which clusters do you want to deploy WizTelemetry Ruler to?
```

#### Step 2: Get Latest Version (if not provided by user)

**MUST do this to get the latest version:**

```bash
kubectl get extensionversions -n kubesphere-system -l kubesphere.io/extension-ref=whizard-telemetry-ruler -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V | tail -1
```

This outputs the latest version (e.g., `1.5.0`). Note this down - you'll use it in the InstallPlan.

#### Step 3: Get AlertManager Host (if configuring sink)

**Only perform this step if you need to configure sink for alert notifications.**

The AlertManager proxy service (`alertmanager-proxy`) is deployed in the host cluster and exposed via NodePort (default port: 31093).

**Step 3.1: Get a host node IP**

```bash
kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'
```

**Step 3.2: Confirm with user**

Ask user to confirm the AlertManager host IP:
```
Detected AlertManager host: <NODE_IP>
Detected AlertManager port: 31093
Alert URL: http://<NODE_IP>:31093/api/v1/alerts

Do you want to use this URL for alert notifications?
```

- If user **confirms** → Use `http://<NODE_IP>:31093/api/v1/alerts` as the sink URL
- If user **provides different URL** → Use the user-specified URL

**Note:** If using WizTelemetry Notification extension, ensure it is installed before installing WizTelemetry Ruler.

### Install WizTelemetry Ruler

**⚠️ IMPORTANT: Complete prerequisite steps BEFORE this step.**

Based on your selections:
- **Target clusters**: User-confirmed cluster names
- **AlertManager URL**: From Step 3 (if configuring sink)

**⚠️ CRITICAL: InstallPlan `metadata.name` MUST be `whizard-telemetry-ruler`. DO NOT use any other name.**

**⚠️ CRITICAL: `config` field is YAML format. You MUST:**
- Use the config structure exactly as shown in the template
- **DO NOT** add configuration fields that are not shown in the template
- **DO NOT** modify the structure or hierarchy

**⚠️ CRITICAL: All placeholders MUST be replaced with actual values. DO NOT leave them as placeholders.**

#### Basic Installation Template (with AlertManager)

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-telemetry-ruler
  namespace: kubesphere-system
spec:
  extension:
    name: whizard-telemetry-ruler
    version: <VERSION>  # From Step 2
  enabled: true
  upgradeStrategy: Manual
  config: |
    whizard-telemetry-ruler:
      config:
        sinks:
          - name: alertmanager
            type: webhook
            config:
              url: http://<ALERT_MANAGER_HOST>:31093/api/v1/alerts  # From Step 3
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

#### Installation with Custom Configuration Template (with all alerting types)

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-telemetry-ruler
  namespace: kubesphere-system
spec:
  extension:
    name: whizard-telemetry-ruler
    version: <VERSION>  # From Step 2
  enabled: true
  upgradeStrategy: Manual
  config: |
    whizard-telemetry-ruler:
      auditingAlerting:
        enabled: true
      eventsAlerting:
        enabled: true
      loggingAlerting:
        enabled: false
      config:
        sinks:
        - name: alertmanager
          type: webhook
          config:
            url: http://<ALERT_MANAGER_HOST>:31093/api/v1/alerts  # From Step 3
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

**Replace placeholders:**
- `<VERSION>`: From Step 2 (e.g., `1.5.0`)
- `<TARGET_CLUSTERS>`: User-confirmed cluster names
- `<ALERT_MANAGER_HOST>`: From Step 3 (auto-detected or user-confirmed node IP)

#### Enable Log Alerting Template

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-telemetry-ruler
  namespace: kubesphere-system
spec:
  extension:
    name: whizard-telemetry-ruler
    version: <VERSION>  # From Step 2
  enabled: true
  upgradeStrategy: Manual
  config: |
    whizard-telemetry-ruler:
      auditingAlerting:
        enabled: true
      eventsAlerting:
        enabled: true
      loggingAlerting:
        enabled: true
      config:
        sinks:
        - name: alertmanager
          type: webhook
          config:
            url: http://<ALERT_MANAGER_HOST>:31093/api/v1/alerts  # From Step 3
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

#### Enable Alerting Persistence Template

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-telemetry-ruler
  namespace: kubesphere-system
spec:
  extension:
    name: whizard-telemetry-ruler
    version: <VERSION>  # From Step 2
  enabled: true
  upgradeStrategy: Manual
  config: |
    global:
       alertingPersistence:
         enabled: true
    whizard-telemetry-ruler:
      config:
        sinks:
        - name: alertmanager
          type: webhook
          config:
            url: http://<ALERT_MANAGER_HOST>:31093/api/v1/alerts  # From Step 3
    alerting-persistence:
      sinks:
        opensearch:
          enabled: true
  clusterScheduling:
    placement:
      clusters:
        - <TARGET_CLUSTERS>
```

## Configuration Parameters

### Alerting Type Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `whizard-telemetry-ruler.auditingAlerting.enabled` | bool | true | Enable auditing alert |
| `whizard-telemetry-ruler.eventsAlerting.enabled` | bool | true | Enable events alert |
| `whizard-telemetry-ruler.loggingAlerting.enabled` | bool | false | Enable log alert |

### Sink Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `whizard-telemetry-ruler.config.sinks[].name` | string | | Sink name |
| `whizard-telemetry-ruler.config.sinks[].type` | string | | Sink type (webhook, etc.) |
| `whizard-telemetry-ruler.config.sinks[].config.url` | string | | Webhook URL |

### Alert Persistence Parameters (Optional)

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `global.alertingPersistence.enabled` | bool | false | Enable alert persistence |
| `alerting-persistence.sinks.opensearch.enabled` | bool | false | Enable OpenSearch sink for alerts |
| `alerting-persistence.sinks.opensearch.ism_policy.enable` | bool | true | Enable ISM policy |
| `alerting-persistence.sinks.opensearch.ism_policy.min_index_age` | string | "7d" | Minimum index retention period |

### Resource Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `whizard-telemetry-ruler.resources.limits.cpu` | string | 2       | ruler CPU limit |
| `whizard-telemetry-ruler.resources.limits.memory` | string | 4Gi     | ruler memory limit |
| `whizard-telemetry-ruler.resources.requests.cpu` | string | 100m    | ruler CPU request |
| `whizard-telemetry-ruler.resources.requests.memory` | string | 20Mi    | ruler memory request |
| `whizard-telemetry-ruler.kubectl.resources.limits.cpu` | string | 100m    | kubectl CPU limit |
| `whizard-telemetry-ruler.kubectl.resources.limits.memory` | string | 256Mi   | kubectl memory limit |
| `whizard-telemetry-ruler.kubectl.resources.requests.cpu` | string | 100m    | kubectl CPU request |
| `whizard-telemetry-ruler.kubectl.resources.requests.memory` | string | 256Mi   | kubectl memory request |

### Node Scheduling Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `whizard-telemetry-ruler.nodeSelector` | map | {} | Node selector |
| `whizard-telemetry-ruler.tolerations` | list | [] | Tolerations |
| `whizard-telemetry-ruler.affinity` | map | {} | Affinity |

## Alerting Rule API

### RuleGroup API (Namespaced)

#### List RuleGroups

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.whizard.io/v1alpha1/namespaces/<namespace>/rulegroups?clusterName=host" \
  -H "X-Remote-User: admin"
```

#### Get RuleGroup

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.whizard.io/v1alpha1/namespaces/<namespace>/rulegroups/<name>?clusterName=host" \
  -H "X-Remote-User: admin"
```

#### Create RuleGroup

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.whizard.io/v1alpha1/namespaces/<namespace>/rulegroups?clusterName=host" \
  -H "X-Remote-User: admin" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "logging.whizard.io/v1alpha1",
    "kind": "RuleGroup",
    "metadata": {
      "name": "<rulegroup-name>",
      "namespace": "<namespace>"
    },
    "spec": {
      "type": "events",
      "rules": [
        {
          "name": "test-rule",
          "desc": "Test rule",
          "enable": true,
          "expr": {
            "kind": "rule",
            "condition": "reason == \"FailedCreatePodSandBox\""
          },
          "alerts": {
            "severity": "warning",
            "message": "Pod sandbox creation failed",
            "labels": {
              "alert": "test"
            },
            "annotations": {
              "summary": "Pod sandbox creation failed"
            }
          }
        }
      ]
    }
  }'
```

#### Update RuleGroup

```bash
curl -X PUT "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.whizard.io/v1alpha1/namespaces/<namespace>/rulegroups/<name>?clusterName=host" \
  -H "X-Remote-User: admin" \
  -H "Content-Type: application/json" \
  -d '<UPDATED_RULEGROUP>'
```

#### Delete RuleGroup

```bash
curl -X DELETE "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.whizard.io/v1alpha1/namespaces/<namespace>/rulegroups/<name>?clusterName=host" \
  -H "X-Remote-User: admin"
```

### ClusterRuleGroup API (Cluster-scoped)

#### List ClusterRuleGroups

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.whizard.io/v1alpha1/clusterrulegroups?clusterName=host" \
  -H "X-Remote-User: admin"
```

#### Get ClusterRuleGroup

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.whizard.io/v1alpha1/clusterrulegroups/<name>?clusterName=host" \
  -H "X-Remote-User: admin"
```

#### Create ClusterRuleGroup

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.whizard.io/v1alpha1/clusterrulegroups?clusterName=host" \
  -H "X-Remote-User: admin" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "logging.whizard.io/v1alpha1",
    "kind": "ClusterRuleGroup",
    "metadata": {
      "name": "<clusterrulegroup-name>"
    },
    "spec": {
      "type": "auditing",
      "rules": [
        {
          "name": "audit-rule",
          "desc": "Audit rule",
          "enable": true,
          "expr": {
            "kind": "rule",
            "condition": "verb == \"delete\""
          },
          "alerts": {
            "severity": "error",
            "message": "Delete operation detected",
            "labels": {
              "type": "audit"
            }
          }
        }
      ]
    }
  }'
```

#### Delete ClusterRuleGroup

```bash
curl -X DELETE "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/logging.whizard.io/v1alpha1/clusterrulegroups/<name>?clusterName=host" \
  -H "X-Remote-User: admin"
```

## Alert Query API

### Query Alerts

Query alerts with filters and time range:

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/events.alerting.wiztelemetry.io/v1alpha1/query" \
  -H "X-Remote-User: admin" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "host",
    "startTime": 1704067200,
    "endTime": 1704153600,
    "from": 0,
    "size": 10,
    "order": "descending",
    "parameters": [
      {
        "field": "severity",
        "operator": "=",
        "value": "error"
      },
      {
        "field": "alertname",
        "operator": "?",
        "values": ["pod*"]
      }
    ]
  }'
```

### Query Statistics

Get alert statistics (overview, histogram, by severity, etc.):

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/events.alerting.wiztelemetry.io/v1alpha1/statistics" \
  -H "X-Remote-User: admin" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "host",
    "statisticsType": 501,
    "startTime": 1704067200,
    "endTime": 1704153600
  }'
```

### Alert Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cluster` | string | Yes | Cluster name (e.g., host, member-1) |
| `startTime` | int64 | No | Start time (Unix timestamp), default: 30 days ago |
| `endTime` | int64 | No | End time (Unix timestamp), default: now |
| `from` | int64 | No | Offset for pagination, default: 0 |
| `size` | int64 | No | Number of results, default: 10 |
| `order` | string | No | Sort order: "ascending" or "descending", default: descending |
| `statisticsType` | int | No | Statistics type (see Statistics Type Values) |
| `parameters` | array | No | Filter parameters |

### Filter Parameter Structure

| Field | Type | Description |
|-------|------|-------------|
| `field` | string | Field name to filter on |
| `operator` | string | Operator (see Filter Operators) |
| `value` | interface{} | Single value for =, !=, >, >=, <, <= |
| `values` | array | Multiple values for In, NotIn, ?, !?, ~, !~ |

### Filter Operators

| Operator | Symbol | Description |
|----------|--------|-------------|
| Equals | = | Exact match |
| NotEquals | != | Not equal |
| Greater | > | Greater than |
| GreaterOrEqual | >= | Greater or equal |
| Less | < | Less than |
| LessOrEqual | <= | Less or equal |
| In | In | In list |
| NotIn | NotIn | Not in list |
| MatchesFuzzy | ? | Fuzzy match (supports * and ?) |
| NotMatchesFuzzy | !? | Not fuzzy match |
| MatchesRegex | ~ | Regex match |
| NotMatchesRegex | !~ | Not regex match |
| Exists | Exists | Field exists |
| NotExists | NotExists | Field does not exist |

### Statistics Type Values

| Type | Value | Description |
|------|-------|-------------|
| StatisticsEventsAlertingNone | 500 | No statistics |
| StatisticsEventsAlertingDateHistogram | 501 | Time histogram |
| StatisticsEventsAlertingOverview | 502 | Overview count |
| StatisticsEventsAlertingByNamespace | 503 | By namespace |
| StatisticsEventsAlertingByRuleGroup | 504 | By rule group |
| StatisticsEventsAlertingByAlertName | 505 | By alert name |
| StatisticsEventsAlertingByAlertType | 506 | By alert type |
| StatisticsEventsAlertingBySeverity | 507 | By severity |

### Available Alert Fields

| Field | Type | Description |
|-------|------|-------------|
| `alertname` | string | Alert name |
| `severity` | string | Alert severity (info, warning, error, critical) |
| `namespace` | string | Namespace |
| `rulegroup` | string | Rule group name |
| `cluster` | string | Cluster name |
| `rulekind` | string | Rule kind (RuleGroup, ClusterRuleGroup) |
| `ruletype` | string | Rule type (events, auditing, logs) |
| `alerttype` | string | Alert type |
| `labels` | string | Alert labels (JSON string) |
| `annotations` | string | Alert annotations (JSON string) |
| `firing` | bool | Is firing |
| `pending` | bool | Is pending |
| `inhibited` | bool | Is inhibited |
| `silenced` | bool | Is silenced |
| `startsat` | int64 | Start timestamp |
| `endsat` | int64 | End timestamp |
| `updatedat` | int64 | Update timestamp |

### Alert Query Examples

#### Query by Severity

```json
{
  "cluster": "host",
  "parameters": [
    {
      "field": "severity",
      "operator": "=",
      "value": "error"
    }
  ],
  "size": 20
}
```

#### Query by Time Range

```json
{
  "cluster": "host",
  "startTime": 1704067200,
  "endTime": 1704153600,
  "size": 100
}
```

#### Query by Namespace and RuleGroup

```json
{
  "cluster": "host",
  "parameters": [
    {
      "field": "namespace",
      "operator": "=",
      "value": "default"
    },
    {
      "field": "rulegroup",
      "operator": "=",
      "value": "my-rule-group"
    }
  ]
}
```

#### Fuzzy Match Alert Name

```json
{
  "cluster": "host",
  "parameters": [
    {
      "field": "alertname",
      "operator": "?",
      "values": ["pod*", "container*"]
    }
  ]
}
```

#### Query with Statistics

```json
{
  "cluster": "host",
  "startTime": 1704067200,
  "endTime": 1704153600,
  "statisticsType": 507
}
```

### RuleGroup API Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `cluster` | string | Cluster name, empty means host cluster |
| `name` | string | Name used for filtering |
| `labelSelector` | string | Label selector used for filtering |
| `status` | string | Filter by enabled status (true or false) |
| `builtin` | string | Filter by builtin status (true or false) |
| `type` | string | Filter by type (logs, events, auditing) |
| `page` | int | Page number |
| `limit` | int | Items per page |
| `orderBy` | string | Sort parameter (e.g., createTime) |
| `ascending` | bool | Sort order |

### Rule Type Values

| Type | Description |
|------|-------------|
| `events` | K8s native events |
| `auditing` | K8s/KubeSphere auditing events |
| `logs` | K8s container logs |

### Alert Severity Values

| Severity | Description |
|----------|-------------|
| `info` | Informational |
| `warning` | Warning |
| `error` | Error |
| `critical` | Critical |

### Rule Expression Kind Values

| Kind | Description |
|------|-------------|
| `rule` | Regular rule with condition |
| `macro` | Macro rule |
| `list` | List rule |
| `alias` | Alias rule |

### Condition Detailed Explanation

The `condition` field is used to filter events/logs/auditing that match specific criteria. It supports various operators and field references.

#### Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equals | `reason == "FailedCreatePodSandBox"` |
| `!=` | Not equals | `verb != "delete"` |
| `=~` | Regex match | `message =~ "error.*failed"` |
| `!~` | Not regex match | `message !~ "debug"` |
| `&&` | AND | `reason == "Failed" && type == "Warning"` |
| `\|\|` | OR | `reason == "Failed" \|\| reason == "Error"` |

#### Available Fields by Type

##### Events Type (`type: events`)

| Field | Type | Description |
|-------|------|-------------|
| `reason` | string | Event reason (e.g., FailedCreatePodSandBox) |
| `type` | string | Event type (Normal, Warning) |
| `involvedObject.kind` | string | Object kind (Pod, Deployment, etc.) |
| `involvedObject.name` | string | Object name |
| `involvedObject.namespace` | string | Object namespace |
| `message` | string | Event message |
| `source` | string | Event source component |
| `count` | int | Event count |

**Events Examples:**
```json
// Alert when pod sandbox creation fails
"condition": "reason == \"FailedCreatePodSandBox\""

// Alert on warning events for specific namespace
"condition": "type == \"Warning\" && involvedObject.namespace == \"default\""

// Alert when event count exceeds threshold
"condition": "count >= 5"
```

##### Auditing Type (`type: auditing`)

| Field | Type | Description |
|-------|------|-------------|
| `verb` | string | HTTP verb (get, post, put, delete, etc.) |
| `user` | string | Username |
| `sourceIPs` | string | Source IP addresses |
| `resource.group` | string | Resource API group |
| `resource.version` | string | Resource API version |
| `resource.resource` | string | Resource type (pods, deployments, etc.) |
| `objectRef.name` | string | Object name |
| `objectRef.namespace` | string | Object namespace |
| `responseStatus.code` | int | HTTP response code |
| `level` | string | Audit level (None, Metadata, Request, RequestResponse) |

**Auditing Examples:**
```json
// Alert on delete operations
"condition": "verb == \"delete\""

// Alert on failed requests (4xx/5xx)
"condition": "responseStatus.code >= 400"

// Alert on specific user activity
"condition": "user == \"admin\" && verb == \"delete\""

// Alert on sensitive resources
"condition": "resource.resource == \"secrets\""
```

##### Logs Type (`type: logs`)

| Field | Type | Description |
|-------|------|-------------|
| `log` | string | Log message content |
| `container` | string | Container name |
| `pod` | string | Pod name |
| `namespace` | string | Namespace name |
| `cluster` | string | Cluster name |

**Logs Examples:**
```json
// Alert on error keyword
"condition": "log contains \"error\""

// Alert on specific container
"condition": "container == \"nginx\""

// Alert on OOM kills
"condition": "log contains \"OOMKilled\""

// Alert on multiple keywords
"condition": "log contains \"failed\" && log contains \"connection\""

// Alert using regex
"condition": "log =~ \"error.*timeout|timeout.*error\""
```

#### Macro Usage

Macros allow reusable expressions:

```json
{
  "expr": {
    "kind": "macro",
    "macro": "high_error_rate"
  }
}
```

#### List Usage

Lists allow grouping values:

```json
{
  "expr": {
    "kind": "list",
    "list": ["error", "warning", "critical"]
  }
}
```

#### Alias Usage

Aliases provide descriptive names for complex expressions:

```json
{
  "expr": {
    "kind": "alias",
    "alias": "Pod_Sandbox_Failure"
  }
}
```

### Sliding Window Alert (Log Alerting)

For log alerting with sliding window:

```json
{
  "name": "log-rate-rule",
  "desc": "Log rate alert",
  "enable": true,
  "expr": {
    "kind": "rule",
    "condition": "log contains \"error\""
  },
  "alerts": {
    "severity": "error",
    "message": "High error log rate"
  },
  "slidingWindow": {
    "windowSize": "5m",
    "slidingInterval": "1m",
    "count": 100
  }
}
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `slidingWindow.windowSize` | string | Window size (e.g., "300ms", "5m") |
| `slidingWindow.slidingInterval` | string | Slide step (must be less than windowSize) |
| `slidingWindow.count` | int | Count threshold to trigger alert |

## Extension Operations

### Check Extension Status

```bash
kubectl get installplan whizard-telemetry-ruler
kubectl get extensionversions -l kubesphere.io/extension-ref=whizard-telemetry-ruler
```

### Uninstall Extension

**Uninstall from all clusters:**

```bash
kubectl delete installplan whizard-telemetry-ruler
```

**Uninstall from specific cluster:**

To remove WizTelemetry Ruler from a specific cluster, update the InstallPlan by removing that cluster from `clusterScheduling.placement.clusters`:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-telemetry-ruler
  namespace: kubesphere-system
spec:
  extension:
    name: whizard-telemetry-ruler
    version: <VERSION>
  enabled: true
  upgradeStrategy: Manual
  clusterScheduling:
    placement:
      clusters:
        - <REMAINING_CLUSTERS>  # Remove the cluster you want to uninstall from
```

## Alert Notification Configuration

To send alerts through WizTelemetry Notification extension, configure the sink URL to point to the `alertmanager-proxy` service.

**Auto-detection (recommended):** Use the command from Step 3 to get the host node IP automatically.

```yaml
whizard-telemetry-ruler:
  config:
    sinks:
    - name: alertmanager
      type: webhook
      config:
        url: http://<ALERT_MANAGER_HOST>:31093/api/v1/alerts
```

- `<ALERT_MANAGER_HOST>`: From Step 3 (auto-detected or user-confirmed)
- Default NodePort: 31093
- Service: `alertmanager-proxy` in `kubesphere-system` namespace
