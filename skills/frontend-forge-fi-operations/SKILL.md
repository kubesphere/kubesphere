---
name: frontend-forge-fi-operations
description: "Operate FrontendIntegration (FI) resources: create, update, enable, disable, delete, and inspect the generated Job, JSBundle, ConfigMap, and manifest. Use this skill when managing FI lifecycle with kubectl, checking FI status, tracing build jobs, reading manifest and source-spec annotations from JSBundle, or debugging missing jobs, missing bundles, incorrect state, or stuck reconciliation."
---

# Frontend Forge FI Operations

## When to use

- Create or update a `FrontendIntegration`
- Enable, disable, or delete an FI
- Inspect FI status and the latest build
- Trace related resources: Job, `JSBundle`, ConfigMap, and manifest
- Debug these cases:
  - no Job triggered
  - no `JSBundle` generated
  - incorrect state or stuck status

## Quick Entry

- Create or update an FI -> see "Create or Update FI"
- Disable an FI -> see "Disable FI"
- Enable an FI -> see "Enable FI"
- Delete an FI -> see "Delete FI"
- Inspect Job / JSBundle / manifest -> see "Inspect Build Output"
- Troubleshoot:
  - no Job -> see "Case 1"
  - Job exists but no Bundle -> see "Case 2"
  - Bundle state is wrong -> see "Case 3"
  - status is stuck -> see "Case 4"

## Preconditions

- The current cluster is reachable through `kubectl`
- These CRDs are installed:
  - `frontendintegrations.frontend-forge.kubesphere.io`
  - `jsbundles.extensions.kubesphere.io`
- `frontend-forge` and `frontend-forge-controller` are running
- The default namespace may be:
  - `extension-frontend-forge`

## Read first

- `references/lifecycle.md`
- `references/inspection.md`
- If available in the current workspace, read a sample FI manifest such as `config/samples/fi-lifecycle-smoke.yaml`
- If workspace docs exist, prefer product docs and deployment manifests over implementation source files

## Resource Model

- `FrontendIntegration`
  - cluster-scoped
  - short name: `fi`
- `JSBundle`
  - cluster-scoped
  - usually named: `fi-<fi-name>`
- ConfigMap
  - default namespace: `extension-frontend-forge`
  - usually named: `<bundle-name>-config`
- Job
  - default namespace: `extension-frontend-forge`
  - triggered during build

## Default Naming

- bundle: `fi-<fi-name>`
- configmap: `<bundle-name>-config`

These are current defaults. Do not hardcode them blindly. If runtime configuration changes, the name or namespace may differ.
When this skill is used outside this repository, treat the live cluster state as the source of truth.

## Common Commands

### Inspect FI

```bash
kubectl get fi <name> -o yaml
kubectl get fi <name> -o jsonpath='{.status}'
```

### Create or Update

```bash
kubectl apply -f <file.yaml>
kubectl apply -f config/samples/fi-lifecycle-smoke.yaml
```

### Disable

```bash
kubectl patch fi <name> --type=merge -p '{"spec":{"enabled":false}}'
```

### Enable

```bash
kubectl patch fi <name> --type=merge -p '{"spec":{"enabled":true}}'
```

### Delete

```bash
kubectl delete fi <name>
```

### Inspect Related Resources

```bash
kubectl get jsbundle <bundle-name> -o yaml
kubectl -n extension-frontend-forge get cm <bundle-name>-config -o yaml
kubectl -n extension-frontend-forge get jobs
kubectl get fi <name> -o jsonpath='{.status.last_build.job_ref.name}{"\n"}'
kubectl get fi <name> -o jsonpath='{.status.bundle_ref.name}{"\n"}'
kubectl get jsbundle <bundle-name> -o jsonpath='{.metadata.annotations.frontend-forge\.io/manifest-content}' | python3 -m json.tool
```

## Recommended Workflows

### 1. Create or Update FI

1. Apply the manifest:

   ```bash
   kubectl apply -f <file.yaml>
   ```

2. Inspect FI status:
   - `.status.phase`
   - `.status.message`
   - `.status.observed_spec_hash`
   - `.status.last_build`

3. Verify:
   - a Job was created
   - `JSBundle` exists
   - `JSBundle.status.state = Available`

### 2. Disable FI

1. Patch:

   ```bash
   kubectl patch fi <name> --type=merge -p '{"spec":{"enabled":false}}'
   ```

2. Expect:
   - `status.message = Disabled`
   - `status.last_build = null`
   - `JSBundle.status.state = Disabled`
   - label `frontend-forge.io/enabled = false`

### 3. Enable FI

1. Patch:

   ```bash
   kubectl patch fi <name> --type=merge -p '{"spec":{"enabled":true}}'
   ```

2. Expect:
   - `JSBundle.status.state = Available`
   - label `frontend-forge.io/enabled = true`
   - `bundle_ref.name` is correct
   - the runtime may reuse an existing bundle or trigger a rebuild, depending on current implementation and cluster state

### 4. Delete FI

```bash
kubectl delete fi <name>
```

Expect:

- FI is removed
- `JSBundle` is removed, or any leftover state is explainable
- ConfigMap is cleaned up, or any leftover state is explainable

### 5. Inspect Build Output

Inspect these `JSBundle.metadata.annotations`:

- `frontend-forge.io/manifest-content`
- `frontend-forge.io/source-spec`
- `frontend-forge.io/source-spec-hash`
- `frontend-forge.io/build-job`

Use them to:

- trace the rendered manifest
- compare the original source spec
- confirm which Job produced the bundle

## Troubleshooting

### Case 1: No Job triggered

Check:

1. `kubectl get fi <name> -o yaml`
   - `phase`
   - `message`
2. whether `spec.enabled` is `true`
3. whether the controller is running:

   ```bash
   kubectl -n extension-frontend-forge get deploy
   ```

4. controller logs

### Case 2: Job exists but no JSBundle

Check:

1. Job status:

   ```bash
   kubectl -n extension-frontend-forge get jobs
   ```

2. Job logs
3. runner errors
4. RBAC and permissions
5. `observed_spec_hash` or stale-spec checks

### Case 3: JSBundle exists but the state is wrong

Check:

1. `JSBundle.status.state`
2. label:
   - `frontend-forge.io/enabled`
3. FI:
   - `bundle_ref`
4. annotations:
   - `source-spec`
   - `manifest-content`

### Case 4: Status is incorrect or stuck

Check:

1. FI:
   - `observed_generation`
   - `observed_spec_hash`
2. controller logs
3. whether reconcile is still progressing

## Validation Checklist

- FI ends in the expected phase:
  - `Succeeded`
  - `Building`
  - `Failed`
  - `Pending`
- `JSBundle` state matches the intended FI state
- source annotations match the actual input
- the build Job matches the current spec

## Pitfalls

- Do not add `-n` when operating on FI; it is cluster-scoped
- Do not trust default naming blindly; if deployment env vars such as `JSBUNDLE_CONFIGMAP_NAMESPACE` changed, inspect runtime config first
- Do not look only at the Job; always check:
  - `FI.status`
  - `JSBundle.status`
  - source annotations
- If admission webhook is disabled, semantic errors may surface at runtime instead of failing during `kubectl apply`

## Escalation

If the issue is still not clear, continue with:

```bash
kubectl -n extension-frontend-forge logs deploy/frontend-forge-controller
kubectl -n extension-frontend-forge logs deploy/frontend-forge
```

Also verify:

- deployment environment variables
- CRD versions
- cluster RBAC
