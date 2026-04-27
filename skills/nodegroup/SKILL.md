---
name: nodegroup
description: NodeGroup operation Skill for the edgewize nodegroup project. Use this whenever the user wants to query, create, update, delete, bind, unbind, or troubleshoot NodeGroup resources, including node binding, namespace binding, workspace binding, and deployment/config inspection for nodegroup.
---

# NodeGroup Operations

## Purpose

Use this skill to perform real `nodegroup`-related operations in the `edgewize-io/nodegroup` environment.

This skill should help with:

- listing and inspecting `NodeGroup`
- creating, updating, patching, and deleting `NodeGroup`
- binding and unbinding nodes
- binding and unbinding namespaces
- binding and unbinding workspaces
- troubleshooting failed or incomplete nodegroup sync
- checking deployment config under `config/nodegroup`

Prefer using the bundled script `scripts/nodegroup_api.py` for authenticated KAPI operations. Use `kubectl` only as a verification or fallback tool when the API path is unavailable.

## Prerequisites

Use the bundled script:

- `scripts/nodegroup_api.py`

Authentication model:

- logs in against `/oauth/token`
- stores token in `~/.kubesphere_token`
- uses `KUBESPHERE_HOST`, `KUBESPHERE_USERNAME`, `KUBESPHERE_PASSWORD`, and `KUBESPHERE_TOKEN`

Setup:

```bash
cd scripts
pip install requests
export KUBESPHERE_HOST="http://<kubesphere-host>"
python nodegroup_api.py login --username admin --password <password>
```

Token helpers:

```bash
python nodegroup_api.py clear-cache
python nodegroup_api.py request GET /kapis/infra.kubesphere.io/v1alpha1/nodegroups
```

## Source of Truth

When you need to confirm how an operation works, read these files first from the nodegroup source repository root:

- `pkg/kapis/infra/v1alpha1/register.go`
- `pkg/kapis/infra/v1alpha1/handler.go`
- `pkg/kapis/infra/v1alpha1/workspace_handler.go`
- `pkg/controller/nodegroup/nodegroup_controller.go`
- `pkg/constants/constants.go`

Do not invent unsupported operations. Base the workflow on the source repository.

## Main Resources

### NodeGroup

Cluster-scoped resource.

Important fields:

- `spec.alias`
- `spec.description`
- `spec.manager`
- `status.state`

## Safe Operating Rules

- Read-only requests: prefer `python nodegroup_api.py ...` query commands.
- Write requests: perform only the exact requested change.
- Do not delete resources unless the user explicitly asks for deletion.
- For destructive or potentially disruptive actions such as node unbind or delete, confirm impact in your response if the request is ambiguous.
- When changing bindings, verify the current state first, then apply the change, then verify the result.

## Preferred Operation Flow

For any write operation, follow this order:

1. Inspect the current resource.
2. Apply the minimal required change.
3. Verify the resulting resource and any affected node, namespace, or workspace.
4. If sync looks wrong, inspect labels, annotations, and controller behavior.

## Core Commands

### 1. Query NodeGroups

```bash
python nodegroup_api.py nodegroup list
python nodegroup_api.py nodegroup get <name>
```

### 2. Create a NodeGroup

```bash
python nodegroup_api.py nodegroup create \
  --name <nodegroup-name> \
  --alias "<alias>" \
  --description "<description>" \
  --manager "<manager>"
python nodegroup_api.py nodegroup get <nodegroup-name>
```

### 3. Update or Patch a NodeGroup

Prefer patching for small changes.

```bash
python nodegroup_api.py nodegroup patch <name> \
  --alias "<new-alias>" \
  --description "<new-description>" \
  --manager "<manager>"
python nodegroup_api.py nodegroup get <name>
```

For unsupported fields or ad hoc testing, use raw request mode.

For `PATCH`, send a JSON Patch array:

```bash
python nodegroup_api.py request PATCH /kapis/infra.kubesphere.io/v1alpha1/nodegroups/<name> '[{"op":"add","path":"/spec/alias","value":"<new-alias>"}]'
```

### 4. Delete a NodeGroup

Only do this when explicitly requested.

```bash
python nodegroup_api.py nodegroup delete <name>
```

If delete hangs, inspect finalizers:

```bash
python nodegroup_api.py nodegroup get <name>
```

Relevant finalizer:

- `finalizers.nodegroups.kubesphere.io`

### 5. Bind or Unbind a Node

Node binding is implemented through nodegroup APIs and reflected with label:

- `apps.edgewize.io/nodegroup=<nodegroup-name>`

Read current state:

```bash
kubectl get node <node-name> --show-labels
kubectl get nodes -l apps.edgewize.io/nodegroup=<nodegroup-name>
```

Operate through the bundled script:

```bash
python nodegroup_api.py bind node --nodegroup <nodegroup-name> --node <node-name>
python nodegroup_api.py unbind node --nodegroup <nodegroup-name> --node <node-name>

# Verify
kubectl get nodes -l apps.edgewize.io/nodegroup=<nodegroup-name>
kubectl get node <node-name> -o yaml
```

### 6. Bind or Unbind a Namespace

Read current state:

```bash
kubectl get ns <namespace> --show-labels
```

Operate through the bundled script:

```bash
python nodegroup_api.py bind namespace --nodegroup <nodegroup-name> --namespace <namespace>
python nodegroup_api.py unbind namespace --nodegroup <nodegroup-name> --namespace <namespace>

# Verify
python nodegroup_api.py nodegroup get <nodegroup-name>
kubectl get ns <namespace> -o yaml
```

### 7. Bind or Unbind a Workspace

```bash
python nodegroup_api.py bind workspace --nodegroup <nodegroup-name> --workspace <workspace>
python nodegroup_api.py unbind workspace --nodegroup <nodegroup-name> --workspace <workspace>
```

## Troubleshooting Checklist

When an operation appears to succeed but state is wrong, check these in order:

1. Does the resource exist?
2. Does it still have a finalizer?
3. Are node labels correct?
4. Are namespace or workspace labels/annotations correct?
5. Did IPPool annotations or labels fail to sync?
6. Is the feature enabled in nodegroup config?

Useful checks:

```bash
python nodegroup_api.py nodegroup get <name>
kubectl get node <node-name> -o yaml
kubectl get ns <namespace> -o yaml
kubectl get cm nodegroup-config -n kubesphere-system -o yaml
kubectl get pods -n kubesphere-system | grep nodegroup
```

Important constants:

- `apps.edgewize.io/nodegroup`
- `apps.edgewize.io/namespace-`
- `nodegroup.infra.kubesphere.io/parent`
- `infra.kubesphere.io/ippool-`

## Deployment and Config

If the user asks how nodegroup is deployed or configured, inspect these files from the nodegroup source repository root:

- `config/nodegroup/templates/nodegroup-apiserver.yml`
- `config/nodegroup/templates/nodegroup-controller-manager.yaml`
- `config/nodegroup/templates/nodegroup-config.yaml`
- `config/nodegroup/values.yaml`

ConfigMap key:

- `nodegroup.yaml`

Notable config sections:

- `scheduling`
- `policy`
- `rbac`
- `ippool`

## API-First Rule

If the user asks for an operation that already has a dedicated nodegroup API, prefer the bundled script and the dedicated nodegroup API semantics over manual label hacking.

Examples:

- node bind/unbind: use `python nodegroup_api.py bind node ...`
- namespace bind/unbind: use `python nodegroup_api.py bind namespace ...`
- workspace bind/unbind: use `python nodegroup_api.py bind workspace ...`
- ad hoc API call: use `python nodegroup_api.py request ...`

Only fall back to direct manifest edits or label repair when the request is explicitly about low-level repair.

## Response Style For This Skill

When using this skill:

- tell the user what operation you are performing
- verify before and after state
- mention any side effects or risks
- if blocked by missing cluster credentials or missing API access, say exactly what is missing
- if the operation is supported in code but cannot be executed in the current environment, provide the exact `nodegroup_api.py` command or API path to run next
