# FrontendIntegration Lifecycle Reference

Primary sources:

- a sample FI manifest, if the current workspace provides one
- deployment or operations docs for the current environment
- the live cluster state from `kubectl get fi ... -o yaml`

## Basics

- `FrontendIntegration` is a cluster-scoped resource
- short name: `fi`
- start from the sample when possible:
  - use a workspace sample such as `config/samples/fi-lifecycle-smoke.yaml` if it exists

## Create

```bash
kubectl apply -f config/samples/fi-lifecycle-smoke.yaml
kubectl get fi fi-lifecycle-smoke -o yaml
kubectl get fi fi-lifecycle-smoke -o jsonpath='{.status.bundle_ref.name}{"\n"}'
```

Fields to inspect:

- `.status.phase`
- `.status.message`
- `.status.observed_spec_hash`
- `.status.observed_generation`
- `.status.last_build`
- `.status.bundle_ref`

## Modify

Prefer editing YAML and re-applying it, or patching a specific field. Example: change the iframe source URL:

```bash
kubectl patch fi fi-lifecycle-smoke --type merge -p \
  '{"spec":{"pages":[{"key":"lifecycle-smoke","type":"iframe","iframe":{"src":"http://example.test/v2"}}]}}'
```

After modification, inspect:

- whether `.status.observed_generation` advanced
- whether `.status.observed_spec_hash` changed
- whether `.status.phase` returned to `Succeeded`
- whether `.status.last_build.job_ref.name` changed

## Disable

```bash
kubectl patch fi fi-lifecycle-smoke --type merge -p '{"spec":{"enabled":false}}'
kubectl get fi fi-lifecycle-smoke -o yaml
```

After disabling, inspect:

- `.status.phase`
- `.status.message`
- `.status.last_build`

## Enable

```bash
kubectl patch fi fi-lifecycle-smoke --type merge -p '{"spec":{"enabled":true}}'
kubectl get fi fi-lifecycle-smoke -o yaml
```

After enabling, inspect:

- `.status.phase`
- `.status.bundle_ref.name`
- `.status.last_build.job_ref.name`
- related `JSBundle.status.state`

## Delete

```bash
kubectl delete fi fi-lifecycle-smoke
```

After deletion, verify again:

- `kubectl get fi fi-lifecycle-smoke`
- `kubectl get jsbundle fi-fi-lifecycle-smoke`
- `kubectl get configmap -n extension-frontend-forge fi-fi-lifecycle-smoke-config`
