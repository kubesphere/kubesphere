---
name: kubesphere-network-extension-operations
description: Operate the KubeSphere network extension. Use when Codex needs to install, upgrade, configure, enable, disable, or inspect the `network` extension; manage Calico `IPPool` resources, namespace bindings, migrations, or network isolation flows; or consult the bundled network extension references in this skill.
---

# KubeSphere Network Extension Operations

Operate the `network` extension and its related resources with `kubectl` plus the bundled references in this skill. Prefer the live cluster state first, then use the copied extension references here to resolve packaging details, API shapes, and product behavior.

## When to Use

- Install, upgrade, enable, disable, or inspect the `network` extension
- Edit `InstallPlan.spec.config` for IPPool or NetworkPolicy feature toggles
- Verify packaging details such as version, images, dependencies, or installation mode
- Manage Calico `IPPool` resources, namespace bindings, occupancy, and migration flows
- Manage cluster, workspace, or project network isolation behavior
- Create, inspect, or troubleshoot Kubernetes `NetworkPolicy` and namespace-level network isolation policies

## Do Not Use

- Manage unrelated extensions
- Assume deprecated `network.kubesphere.io` IPPool CRDs are still the CRUD source of truth
- Assume a non-Calico IPPool backend; the extension values currently support only `calico`
- Patch workspace, namespace, or IPPool binding metadata without reading the current live object first

## Read First

Read only the references needed for the current request:

- Product behavior and upgrade caveats: [references/README_zh.md](references/README_zh.md)
- IPPool and NetworkPolicy API flows: [references/api_doc.md](references/api_doc.md)
- Extension defaults and feature toggles: [references/values.yaml](references/values.yaml)
- Packaging facts such as version, dependencies, images, and installation mode: [references/extension.yaml](references/extension.yaml)
- Full endpoint schema details only when request or response fields matter: [references/swagger.yaml](references/swagger.yaml)

Useful `rg` patterns for the larger references:

```bash
rg -n "ippool|networkpol|isolate" skills/kubesphere-network-extension-operations/references/api_doc.md
rg -n "^  /(kapis|apis)/.*(ippool|networkpol)" skills/kubesphere-network-extension-operations/references/swagger.yaml
```

## Fixed Facts

- Extension name: `network`
- InstallPlan name: `network`
- Installation mode from packaging: `Multicluster`
- Current packaged version in the copied reference: `1.3.0`
- Supported IPPool backend in values: `global.ippool.type=calico`
- Main feature toggles:
  - `global.ippool.enable`
  - `global.ippool.webhook`
  - `global.networkPolicy.enable`

## Workflow

1. Read the matching bundled references for the request.
2. Inspect the current live extension, InstallPlan, and relevant cluster resources.
3. Apply the smallest change that satisfies the request.
4. Re-read the changed resources and verify the post-change state.

## Preflight Checks

- `kubectl` must already be configured for the target cluster.
- Before any mutation, confirm the live extension state:

```bash
kubectl get extension network
kubectl get installplan network -o yaml
kubectl get extensionversion.kubesphere.io -l kubesphere.io/extension-ref=network
```

- Before any IPPool CRUD, confirm the Calico CRD exists:

```bash
kubectl get crd ippools.crd.projectcalico.org
```

- Before any network isolation mutation, inspect one live object of the same type first:

```bash
kubectl get workspace <workspace-name> -o yaml
kubectl get namespace <namespace-name> -o yaml
```

## Extension Management

Use fixed resource names:

- extension: `network`
- InstallPlan: `network`

When creating or updating an InstallPlan:

- Use the exact version requested by the user.
- Keep `metadata.name` and `spec.extension.name` both equal to `network`.
- Use `upgradeStrategy: Manual`.
- Add `clusterScheduling` only when the user explicitly needs agent placement in member clusters.
- Omit `spec.config` unless the user asked for non-default settings.
- On upgrade, inspect the current config first and remove or update any pinned old image tags. The bundled README warns that stale tags can block new images from being deployed.

Minimal default example:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: network
spec:
  enabled: true
  extension:
    name: network
    version: <exact-version>
  upgradeStrategy: Manual
```

Optional config snippet when the user asks to customize features:

```yaml
config: |
  global:
    ippool:
      enable: true
      type: calico
      webhook: true
    networkPolicy:
      enable: true
```

Prefer this inspection flow before install or upgrade:

```bash
kubectl get extension network -o yaml
kubectl get extensionversion network-<exact-version> -o yaml
kubectl get installplan network -o yaml
```

## IPPool Operations

Treat Calico IPPools as the CRUD source of truth:

- Create, update, delete: `/apis/crd.projectcalico.org/v1/ippools`
- IP usage and occupancy details:
  - `/kapis/network.kubesphere.io/v1alpha2/ippools`
  - `/kapis/network.kubesphere.io/v1alpha2/ippools/{name}`
- Migration:
  - `POST /kapis/network.kubesphere.io/v1alpha2/ippoolmigrations`
  - `GET /kapis/network.kubesphere.io/v1alpha2/ippools/{name}/migrate`
- Namespace available IPPools:
  - `GET /kapis/network.kubesphere.io/v1alpha2/namespaces/{namespace}/ippools`

Operational rules:

- Do not recreate old KubeSphere-managed IPPool CRDs from `network.kubesphere.io`.
- Before binding or unbinding a namespace to an IPPool, inspect a currently bound namespace and preserve the live label or annotation format already used by the cluster.
- Before migrating an IPPool, inspect the source IPPool, bound namespaces, and current pod allocations first.
- Use the examples in [references/api_doc.md](references/api_doc.md) for bound-namespace and pod-occupancy query shapes instead of inventing new label selectors.

Suggested live inspection commands:

```bash
kubectl get ippools.crd.projectcalico.org
kubectl get ippools.crd.projectcalico.org <ippool-name> -o yaml
kubectl get namespace <namespace-name> -o yaml
kubectl get pods -A -o wide
```

## NetworkPolicy Operations

### Cluster-Scope and Namespace-Scope Kubernetes NetworkPolicy

Use Kubernetes `networking.k8s.io/v1` endpoints for standard `NetworkPolicy` CRUD:

- List:
  - `/kapis/networking.k8s.io/v1/networkpolicies`
  - `/kapis/networking.k8s.io/v1/namespaces/{namespace}/networkpolicies`
- Create:
  - `POST /kapis/networking.k8s.io/v1/namespaces/{namespace}/networkpolicies`
- Delete:
  - `DELETE /kapis/networking.k8s.io/v1/namespaces/{namespace}/networkpolicies/{name}`

### Workspace and Project Isolation

The bundled API document uses inconsistent prose for the annotation key: its status-check text mentions `kubesphere.io/workspace-isolate`, while its patch examples use `kubesphere.io/network-isolate`.

Follow this rule:

- Inspect the live workspace or namespace annotations before patching.
- Unless the live cluster proves otherwise, use the example payload key `kubesphere.io/network-isolate: enabled`.

Typical patch body:

```json
{
  "metadata": {
    "annotations": {
      "kubesphere.io/network-isolate": "enabled"
    }
  }
}
```

### Namespace-Level Network Isolation Policies

Use the KubeSphere API for project-specific isolation policies:

- List:
  - `GET /kapis/network.kubesphere.io/v1alpha1/namespaces/{namespace}/namespacenetworkpolicies`
- Create:
  - `POST /kapis/network.kubesphere.io/v1alpha1/namespaces/{namespace}/namespacenetworkpolicies`
- Update:
  - `PUT /kapis/network.kubesphere.io/v1alpha1/namespaces/{namespace}/namespacenetworkpolicies/{name}`
- Delete:
  - `DELETE /kapis/network.kubesphere.io/v1alpha1/namespaces/{namespace}/namespacenetworkpolicies/{name}`

When creating or filtering these policies, preserve these labels exactly:

- `kubesphere.io/policy-type=egress`
- `kubesphere.io/policy-type=ingress`
- `kubesphere.io/policy-traffic=inside`
- `kubesphere.io/policy-traffic=outside`

## Troubleshooting

When the extension install or upgrade is stuck, gather state in this order:

```bash
kubectl describe extension network
kubectl describe installplan network
kubectl get installplan network -o jsonpath='{.status.targetNamespace}{"\n"}'
kubectl get pods,svc -n <target-namespace>
kubectl get jobs -A | rg 'helm-upgrade-network|network'
kubectl get pods -n kubesphere-system
```

If the InstallPlan already points to a target namespace, inspect the Helm job pod logs and any extension pods in that namespace before changing manifests again.

## Rules

- Keep answers grounded in the bundled references and current live resource state.
- Prefer exact version numbers and explicit resource reads over assumptions such as "latest" or "default" without verification.
- Surface ambiguity when the copied docs and the live cluster object shape disagree.
- Treat `extension`, `installplan`, Calico IPPools, and the live workspace or namespace objects as the source of truth.
