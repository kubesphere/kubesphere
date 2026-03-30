---
name: kubesphere-cluster-management
description: KubeSphere cluster query Skill (read-only). Use when user requests to view cluster list, cluster status, cluster details, or cluster version info. Do not perform any write operations (create, delete, update). Use KubeSphere Console for management operations.
---

# KubeSphere Cluster Query

> **⚠️ Security**: Read-only only. **Never** use kubectl to create, update, or delete cluster resources. Use KubeSphere console instead.

## When to Use

List or get KubeSphere cluster information: list clusters, get details, check status, identify host/member.

## Commands

```bash
# List clusters
kubectl get cluster

# Get cluster details
kubectl get cluster <name> -o yaml

# Get cluster summary (name, connection type, version, status)
kubectl get cluster -o go-template='
{{printf "%-12s %-12s %-12s %-15s %-13s %-12s\n" "NAME" "CONN_TYPE" "KS_VERSION" "K8S_VERSION" "ClusterReady" "KSCORE_READY"}}
{{range .items}}{{printf "%-12s %-12s %-12s %-15s" .metadata.name .spec.connection.type .status.kubeSphereVersion .status.kubernetesVersion}}{{range .status.conditions}}{{if eq .type "Ready"}}{{printf "%-13s" .status}}{{end}}{{end}}{{range .status.conditions}}{{if eq .type "KSCoreReady"}}{{printf "%-12s" .status}}{{end}}{{end}}{{"\n"}}{{end}}'
```

## Out of Scope

Cluster installation/upgrade/remove, any write operations. Use KubeSphere console.
