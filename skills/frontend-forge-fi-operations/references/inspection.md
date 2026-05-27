# FI Runtime Inspection Reference

Primary sources:

- FI, JSBundle, Job, and ConfigMap objects from the live cluster
- deployment manifests and environment variables for the running controller and runner
- workspace docs, if the current repo includes operations or architecture notes

## Default Naming Rules

- default bundle name:
  - `fi-<fi-name>`
- default Job name prefix:
  - `fi-<fi-name>-build-<hash8>`
- default ConfigMap name:
  - `<bundle-name>-config`

These are common defaults. If runtime environment variables changed at deployment time, inspect controller configuration first. When using this skill outside this repository, prefer observed cluster object names over assumed defaults.

## Query Related Resources

Inspect FI:

```bash
kubectl get fi <fi-name> -o yaml
```

Inspect Job:

```bash
kubectl get jobs -n extension-frontend-forge -l frontend-forge.io/fi-name=<fi-name>
kubectl get fi <fi-name> -o jsonpath='{.status.last_build.job_ref.name}{"\n"}'
kubectl describe job -n extension-frontend-forge <job-name>
kubectl logs -n extension-frontend-forge job/<job-name>
```

Inspect JSBundle:

```bash
kubectl get fi <fi-name> -o jsonpath='{.status.bundle_ref.name}{"\n"}'
kubectl get jsbundle fi-<fi-name> -o yaml
```

Inspect ConfigMap:

```bash
kubectl get configmap -n extension-frontend-forge fi-<fi-name>-config -o yaml
```

## Important Labels and Annotations

Common labels:

- `frontend-forge.io/fi-name`
- `frontend-forge.io/spec-hash`
- `frontend-forge.io/manifest-hash`
- `frontend-forge.io/enabled`

Common annotations:

- `frontend-forge.io/build-job`
- `frontend-forge.io/manifest-hash`
- `frontend-forge.io/manifest-content`
- `frontend-forge.io/source-spec`
- `frontend-forge.io/source-spec-hash`
- `frontend-forge.io/source-generation`

## Inspect Manifest and Source Spec Directly

Inspect the manifest:

```bash
kubectl get jsbundle fi-<fi-name> -o jsonpath='{.metadata.annotations.frontend-forge\.io/manifest-content}'
kubectl get jsbundle fi-<fi-name> -o jsonpath='{.metadata.annotations.frontend-forge\.io/manifest-content}' | python3 -m json.tool
```

Inspect the source spec:

```bash
kubectl get jsbundle fi-<fi-name> -o jsonpath='{.metadata.annotations.frontend-forge\.io/source-spec}'
```

Inspect the build Job:

```bash
kubectl get jsbundle fi-<fi-name> -o jsonpath='{.metadata.annotations.frontend-forge\.io/build-job}'
```

## Key Status Fields

FI status:

- `.status.phase`
- `.status.observed_spec_hash`
- `.status.observed_manifest_hash`
- `.status.observed_generation`
- `.status.last_build.job_ref.name`
- `.status.last_error.message`
- `.status.bundle_ref.name`
- `.status.message`

JSBundle status:

- `.status.state`
- `.status.link`

## Troubleshooting Order

1. Check FI `status` first
2. Then check `bundle_ref.name` and `last_build.job_ref.name`
3. Then inspect Job logs, JSBundle annotations, and ConfigMap content
4. If FI did not create a Job, focus on whether webhook or controller rejected or skipped the spec
