---
name: whizard-notification
description: Use when working with WizTelemetry Notification extension for KubeSphere, including installation, configuration, troubleshooting, notification channel setup, alert routing, and silence management.
---

# WizTelemetry Notification

## Overview

WizTelemetry Notification is the notification component of the KubeSphere observability platform. It receives alerts, cloud events, and audit logs in a multi-tenant Kubernetes environment and distributes notifications to different channels based on tenant labels (e.g., namespace).


## Architecture

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                             Alertmanager                                     │
│      Prometheus  ──▶ metrics alert  ──▶ notification-manager-svc:19093       │
│      K8s Events  ──▶ events alert   ──▶ notification-manager-svc:19093       │
│      Auditing    ──▶ auditing alert ──▶ notification-manager-svc:19093       │
│      Logging     ──▶ logging alert  ──▶ notification-manager-svc:19093       │
└──────────────────────────────────────────────────────────────────────────────┘
                                    │
                    notification-manager (:19093)
                                    │
         ┌──────────────────────────┼──────────────────────────┐
         │                          │                          │
    [Silence Stage]          [Route Stage]             [Filter Stage]
         │                          │                          │
         ▼                          ▼                          ▼
 [Match Silence CRs]         [Match Router CRs]      [alertSelector filtering]
                                 │
                          [Aggregation Stage]
                                 │
                          [Notify Stage]
                                 │
         ┌────────────────────────┼────────────────────────────┐
         │                        │                            │
    Email/Slack              DingTalk/Feishu            WeChat/Webhook
         │
         ▼
  [History Webhook]
         │
         ▼
┌─────────────────────┐
│  Vector Aggregator  │
└────────┬────────────┘
         │
         │
         ▼
     OpenSearch
```

## Components

| Component | Description | Default |
|-----------|-------------|---------|
| **notification-manager** | Core notification management, receives alerts, dispatches to channels | true |
| **alertmanager** (v0.27.0) | Alert routing, deduplication, aggregation, silencing | true |
| **alertmanager-proxy** | Alertmanager API proxy, NodePort 31093 | true |
| **notification-history** | Stores sent notifications via Vector to OpenSearch | true |

## Dependencies

| Dependency | Type | Required |
|------------|------|----------|
| **whizard-telemetry** | extension | Required |
| **vector** (WizTelemetry Data Pipeline) | extension | Required only notification history is enabled |
| **opensearch** | extension | Required only notification history is enabled |

## CRDs

All CRDs belong to `notification.kubesphere.io/v2beta2`, scope is Cluster.

| CRD | Kind | Description |
|-----|------|-------------|
| `notificationmanagers.notification.kubesphere.io` | NotificationManager | Notification manager deployment spec (image, replicas, resources, sidecar, receiver config) |
| `receivers.notification.kubesphere.io` | Receiver | Tenant-level notification receiver, references Config |
| `configs.notification.kubesphere.io` | Config | Credential references (via Secret selectors), independent per channel |
| `routers.notification.kubesphere.io` | Router | Route alerts to receivers based on label matchers |
| `silences.notification.kubesphere.io` | Silence | Time-bounded silencing, supports matchers |

### Supported Notification Channels (11 channels)

| Channel | Receiver Field | Config Field | Description |
|---------|---------------|--------------|-------------|
| Email | `spec.email` | `spec.email` | SMTP with TLS/STARTTLS support |
| Slack | `spec.slack` | `spec.slack` | Slack webhook + token |
| DingTalk | `spec.dingtalk` | `spec.dingtalk` | ChatBot + group conversation |
| WeChat | `spec.wechat` | `spec.wechat` | ChatBot + group conversation |
| Feishu | `spec.feishu` | `spec.feishu` | ChatBot + group conversation |
| Webhook | `spec.webhook` | `spec.webhook` | Generic HTTP / HTTPS POST |

### Config → Receiver Association Mechanism

```
Config (credentials) --labels--> Receiver (via label selector) --labels--> Router (route matching)
                        │
                        └── namespace for multi-tenant isolation
```

Config's `labels` are used by Receiver's `*ConfigSelector`. Receiver's `labels` are used by Router's `receiverSelector`.

## Installation

### Prerequisites

**REQUIRED: Complete all steps in order before generating InstallPlan.**

#### Step 1: Get Latest Version (if not provided by user)

```bash
kubectl get extensionversions -l kubesphere.io/extension-ref=whizard-notification -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V | tail -1
```

This outputs the latest version (e.g., `2.7.0`). Note this down.


#### Step 2: Create InstallPlan

**CRITICAL:**
- `metadata.name` MUST be `whizard-notification`
- `config` field is YAML format
- DO NOT add fields not shown in the template
- Replace all placeholders with actual values

**Basic installation:**

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-notification
spec:
  extension:
    name: whizard-notification
    version: <VERSION>  # From Step 1
  enabled: true
  upgradeStrategy: Manual
```

**Disable notification history:**

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: whizard-notification
spec:
  extension:
    name: whizard-notification
    version: <VERSION>
  enabled: true
  upgradeStrategy: Manual
  config: |
    notification-history:
      enabled: false
```

**Custom index format (monthly):**

```yaml
notification-history:
  sinks:
    opensearch:
      index:
        prefix: "{{ .cluster }}-notification-history"
        timestring: "%Y.%m"
```

## Configuration Parameters

### Notification History

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `notification-history.enabled` | bool | true | Enable notification history storage |
| `notification-history.sinks.opensearch.enabled` | bool | true | Enable OpenSearch sink |
| `notification-history.sinks.opensearch.index.prefix` | string | `{{ .cluster }}-notification-history` | Index prefix, supports template |
| `notification-history.sinks.opensearch.index.timestring` | string | `%Y.%m.%d` | strftime format |
| `notification-history.ism_policy.enable` | bool | false | Enable ISM policy |
| `notification-history.ism_policy.min_index_age` | string | `7d` | Minimum index retention |

### Alertmanager

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `alertmanager.replicaCount` | int | 1 | Replica count |
| `alertmanager.replicaAutoSet` | bool | true | Auto set replicas by node count (<3=1, >=3=3) |
| `alertmanager.service.nodePort` | int | 31093 | NodePort (alertmanager-proxy) |
| `alertmanager.config.group_wait` | string | 30s | Alert aggregation wait time |
| `alertmanager.config.repeat_interval` | string | 12h | Repeat alert interval |
| `alertmanager.config.route.receiver` | string | Default | Default receiver |
| `alertmanager.config.inhibit_rules` | list | see below | Inhibition rules (critical→warning→info) |

**Default inhibit_rules:**
- `critical` alerts silence `warning` and `info` with same labels
- `warning` alerts silence `info` with same labels
- `InfoInhibitor` silences all `info` level alerts

### Notification Manager

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `notification-manager.notificationmanager.sidecar.tenant.resources.limits.cpu` | string | 200m | Sidecar CPU limit |
| `notification-manager.notificationmanager.sidecar.tenant.resources.limits.memory` | string | 512Mi | Sidecar memory limit |
| `notification-manager.notificationmanager.sidecar.tenant.resources.requests.cpu` | string | 200m | Sidecar CPU request |
| `notification-manager.notificationmanager.sidecar.tenant.resources.requests.memory` | string | 256Mi | Sidecar memory request |
| `notification-manager.operator.containers.operator.resources.limits.cpu` | string | 50m | notification manager operator CPU limit |
| `notification-manager.operator.containers.operator.resources.limits.memory` | string | 50Mi | notification manager operator memory limit |
| `notification-manager.operator.containers.operator.resources.requests.cpu` | string | 5m | notification manager operator CPU request |
| `notification-manager.operator.containers.operator.resources.requests.memory` | string | 20Mi | notification manager operator memory request |
| `notification-manager.notificationmanager.resources.limits.cpu` | string | 500m | notification manager CPU limit |
| `notification-manager.notificationmanager.resources.limits.memory` | string | 500Mi | notification manager memory limit |
| `notification-manager.notificationmanager.resources.requests.cpu` | string | 5m | notification manager CPU request |
| `notification-manager.notificationmanager.resources.requests.memory` | string | 20Mi | notification manager memory request |
| `notification-manager.notificationmanager.replicas` | int | 1 | notification manager replicas |

### Alertmanager-Proxy

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `alertmanager-proxy.service.nodePort` | int | 31093 | NodePort |
| `alertmanager-proxy.resources.limits.cpu` | string | 500m | CPU limit |
| `alertmanager-proxy.resources.limits.memory` | string | 500Mi | Memory limit |
| `alertmanager-proxy.resources.requests.cpu` | string | 100m | CPU request |
| `alertmanager-proxy.resources.requests.memory` | string | 100Mi | Memory request |

## Notification API

All APIs are at `/kapis/notification.kubesphere.io/v2beta2/`, routed through `whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80`. No authentication token is required.

### API Patterns

The API supports two levels of resources:

| Level | URL Pattern | Resources |
|-------|-------------|-----------|
| **Global** | `/{resources}` | `notificationmanagers`, `configs`, `receivers`, `routers`, `silences`, `secrets`, `configmaps` |
| **Tenant** | `/users/{user}/{resources}` | `configs`, `receivers`, `silences`, `secrets`, `configmaps` (NOT `notificationmanagers` or `routers`) |

Supported HTTP methods: `GET` (list/detail), `POST` (create), `PUT` (update), `PATCH` (patch), `DELETE` (delete).

Query parameters for list endpoints: `name`, `labelSelector`, `type`, `page`, `limit`, `ascending`, `orderBy`.


### List Global Receivers

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/receivers" \
```

### List Global Receivers (filtered by type)

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/receivers?type=email" \
```

### List Tenant Receivers

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/users/admin/receivers" \
```

### Create Receiver — Email (Global)

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/receivers" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "notification.kubesphere.io/v2beta2",
    "kind": "Receiver",
    "metadata": {
      "name": "email-receiver",
      "labels": {
        "type": "default"
      }
    },
    "spec": {
      "email": {
        "emailConfigSelector": {
          "matchLabels": { "type": "default" }
        },
        "to": ["user@example.com"],
        "tmplType": "html"
      }
    }
  }'
```

### Create Receiver — Email (Tenant)

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/users/admin/receivers" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "notification.kubesphere.io/v2beta2",
    "kind": "Receiver",
    "metadata": {
      "name": "email-receiver",
      "labels": {
        "type": "tenant",
        "user": "admin"
      }
    },
    "spec": {
      "email": {
        "emailConfigSelector": {
          "matchLabels": { "type": "default" }
        },
        "to": ["user@example.com"],
        "tmplType": "html"
      }
    }
  }'
```

### Create Receiver — Slack (Tenant)

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/users/admin/receivers" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "notification.kubesphere.io/v2beta2",
    "kind": "Receiver",
    "metadata": {
      "name": "slack-receiver",
      "labels": { "type": "tenant", "user": "admin" }
    },
    "spec": {
      "slack": {
        "slackConfigSelector": {
          "matchLabels": { "type": "default" }
        },
        "channels": ["#alerts"],
        "alertSelector": {
          "matchLabels": { "severity": "critical" }
        }
      }
    }
  }'
```

### Create Receiver — DingTalk ChatBot (Global)

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/receivers" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "notification.kubesphere.io/v2beta2",
    "kind": "Receiver",
    "metadata": {
      "name": "dingtalk-receiver",
      "labels": { "type": "global" }
    },
    "spec": {
      "dingtalk": {
        "dingtalkConfigSelector": {
          "matchLabels": { "type": "default" }
        },
        "chatbot": {
          "webhook": {
            "value": "<CHATBOT_WEBHOOK>"
          },
          "secret": {
            "valueFrom": {
              "secretKeyRef": {
                "name": "dingtalk-secret",
                "key": "secret"
              }
            }
          },
          "keywords": ["Alert", "alert"],
          "atAll": false
        },
        "alertSelector": {}
      }
    }
  }'
```

### Create Receiver — Feishu (Global)

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/receivers" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "notification.kubesphere.io/v2beta2",
    "kind": "Receiver",
    "metadata": {
      "name": "feishu-receiver",
      "labels": { "type": "global" }
    },
    "spec": {
      "feishu": {
        "feishuConfigSelector": {
          "matchLabels": { "type": "default" }
        },
        "chatbot": {
          "webhook": {
            "value": "<FEISHU_WEBHOOK>"
          },
          "secret": {
            "valueFrom": {
              "secretKeyRef": {
                "name": "feishu-secret",
                "key": "secret"
              }
            }
          },
          "keywords": ["Alert"]
        },
        "user": ["<USER_ID>"],
        "department": ["<DEPT_ID>"]
      }
    }
  }'
```

### Create Receiver — Webhook (Tenant)

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/users/admin/receivers" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "notification.kubesphere.io/v2beta2",
    "kind": "Receiver",
    "metadata": {
      "name": "webhook-receiver",
      "labels": { "type": "tenant", "user": "admin" }
    },
    "spec": {
      "webhook": {
        "webhookConfigSelector": {
          "matchLabels": { "type": "default" }
        },
        "url": {
          "value": "https://example.com/webhook"
        },
        "tmplType": "json"
      }
    }
  }'
```

### Update Receiver

```bash
curl -X PUT "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/receivers/email-receiver" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "notification.kubesphere.io/v2beta2",
    "kind": "Receiver",
    "metadata": {
      "name": "email-receiver",
      "labels": { "type": "default" }
    },
    "spec": {
      "email": {
        "emailConfigSelector": {
          "matchLabels": { "type": "default" }
        },
        "to": ["user@example.com", "admin@example.com"],
        "tmplType": "html"
      }
    }
  }'
```

### Patch Receiver

```bash
curl -X PATCH "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/receivers/email-receiver" \
  -H "Content-Type: application/json" \
  -d '{"spec": {"email": {"to": ["new@example.com"]}}}'
```

### Delete Receiver

```bash
curl -X DELETE "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/receivers/email-receiver"
```

### List Silences (Global)

```bash
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/silences"
```

### Create Silence (Global)

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/silences" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "notification.kubesphere.io/v2beta2",
    "kind": "Silence",
    "metadata": {
      "name": "maintenance-window"
    },
    "spec": {
      "matchers": [
        {"name": "alertname", "value": "KubeNodeReady"},
        {"name": "namespace", "value": "default", "operator": "Equal"}
      ],
      "startAt": "2024-01-01T00:00:00Z",
      "endAt": "2024-01-02T00:00:00Z",
      "createdBy": "admin",
      "comment": "Scheduled maintenance"
    }
  }'
```

### Create Silence (Tenant)

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/users/admin/silences" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "notification.kubesphere.io/v2beta2",
    "kind": "Silence",
    "metadata": {
      "name": "tenant-maintenance",
      "labels": { "type": "tenant", "user": "admin" }
    },
    "spec": {
      "matchers": [
        {"name": "alertname", "value": ".*"}
      ],
      "startAt": "2024-01-01T00:00:00Z",
      "endAt": "2024-01-02T00:00:00Z",
      "createdBy": "admin",
      "comment": "Tenant maintenance window"
    }
  }'
```

### Delete Silence

```bash
curl -X DELETE "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/silences/maintenance-window"
```

### Create Router (Global only)

```bash
curl -X POST "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/routers" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "notification.kubesphere.io/v2beta2",
    "kind": "Router",
    "metadata": {
      "name": "critical-alerts-router"
    },
    "spec": {
      "receivers": ["email-receiver", "slack-receiver"],
      "alertSelector": {
        "matchLabels": { "severity": "critical" }
      }
    }
  }'
```

### Search Notification History

```bash
# Basic search
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/notifications/search?size=10"

# Search with filters
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/notifications/search?alertname=KubeNodeReady&severity=critical&namespace=default&size=20&sort=startTime&order=desc"

# Search with time range
curl -X GET "http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:80/kapis/notification.kubesphere.io/v2beta2/notifications/search?start_time=1704067200&end_time=1704153600&size=50"
```

**Notification Search Query Parameters:**

| Parameter | Description |
|-----------|-------------|
| `status` | Alert status: `firing`, `resolved` (comma-separated) |
| `alertname` | Alert name (comma-separated, exact match) |
| `alertname_fuzzy` | Alert name fuzzy match |
| `alerttype` | Alert type (comma-separated) |
| `alerttype_fuzzy` | Alert type fuzzy match |
| `severity` | Severity level (comma-separated) |
| `severity_fuzzy` | Severity level fuzzy match |
| `namespace` | Namespace (comma-separated) |
| `namespace_fuzzy` | Namespace fuzzy match |
| `message_fuzzy` | Alert message fuzzy match |
| `start_time` | Start time (Unix timestamp, seconds) |
| `end_time` | End time (Unix timestamp, seconds) |
| `sort` | Sort field (e.g., `startTime`) |
| `order` | Sort order: `asc` or `desc` (default: desc) |
| `from` | Offset from result set (default: 0) |
| `size` | Number of results (default: 10, max: 100) |
| `cluster` | Cluster name |

## Extension Operations

### Check Extension Status

```bash
kubectl get installplan whizard-notification
kubectl get extension whizard-notification
kubectl get extensionversion -l kubesphere.io/extension-ref=whizard-notification
```

### Check Pod Status

```bash
# notification-manager operator + deployment
kubectl get pods -n kubesphere-monitoring-system -l notification-manager=notification-manager

# alertmanager
kubectl get pods -n kubesphere-monitoring-system -l app.kubernetes.io/name=alertmanager

# alertmanager-proxy
kubectl get pods -n kubesphere-monitoring-system -l name=alertmanager-proxy
```

### View Logs

```bash
# notification-manager main process
kubectl logs -n kubesphere-monitoring-system -l notification-manager=notification-manager -c notification-manager --tail=100

# tenant sidecar
kubectl logs -n kubesphere-monitoring-system -l notification-manager=notification-manager -c tenant --tail=100

# operator
kubectl logs -n kubesphere-monitoring-system -l control-plane=controller-manager --tail=100

# alertmanager
kubectl logs -n kubesphere-monitoring-system -l app.kubernetes.io/name=alertmanager --tail=100
```

### Uninstall Extension

```bash
kubectl delete installplan whizard-notification
```

## Troubleshooting

### Issue 1: Notifications Not Sent

**Diagnosis steps:**

```bash
# 1. Check if notification-manager is running
kubectl get pods -n kubesphere-monitoring-system -l notification-manager=notification-manager

# 2. Check if alertmanager receives alerts
kubectl logs -n kubesphere-monitoring-system -l app.kubernetes.io/name=alertmanager --tail=50

# 3. Check notification-manager logs
kubectl logs -n kubesphere-monitoring-system -l notification-manager=notification-manager -c notification-manager --tail=100 | grep -i error

# 4. Check if Receiver and Config exist
kubectl get receivers,configs -n kubesphere-monitoring-system

# 5. Check if silenced by Silence CR
kubectl get silences
```

### Issue 2: Notification History Empty

```bash
# 1. Check if vector aggregator is running
kubectl get pods -n kubesphere-logging-system -l app.kubernetes.io/name=vector-aggregator

# 2. Check if vector-sinks secret exists
kubectl get secret vector-sinks -n kubesphere-logging-system

# 3. Check OpenSearch indices exist
# (requires OpenSearch API or Kibana)
```

### Issue 3: Tenant Sidecar Not Responding

```bash
# Check sidecar logs
kubectl logs -n kubesphere-monitoring-system -l notification-manager=notification-manager -c tenant --tail=50

# Test sidecar API
kubectl exec -n kubesphere-monitoring-system deploy/notification-manager -- \
  curl -s "http://localhost:19094/api/v2/tenant?namespace=default"
```

### Issue 4: Upgrade Still Using Old Images

```bash
# Check image tag config; manually set old tags block upgrades
kubectl get installplan whizard-notification -n kubesphere-system -o yaml | grep -i image

# Remove or update manually set image.tag in InstallPlan config
```

## Important Notes

1. **Dependency order**: Install `vector`, and `opensearch` first if notification history is enabled
2. **OpenSearch**: Notification history sink credentials are auto-loaded from `vector-sinks` Secret
3. **Multi-tenancy**: Tenant sidecar (port 19094) resolves namespace → authorized user mappings
4. **Image tags**: Before upgrading, check and remove manually set `image.tag` to avoid pulling stale images
