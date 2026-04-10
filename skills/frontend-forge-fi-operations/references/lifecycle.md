# FrontendIntegration Lifecycle Reference

Extension must be installed and enabled before any FI lifecycle action. See `SKILL.md` Preflight.

## Resource Basics

- `FrontendIntegration` is cluster-scoped
- short name: `fi`

## Create From YAML

Use complete `FrontendIntegration` YAML only.

```bash
kubectl apply -f <fi.yaml>
kubectl get fi <fi-name> -o yaml
kubectl get fi <fi-name> -o jsonpath='{.status.phase}{"\n"}'
kubectl get fi <fi-name> -o jsonpath='{.status.message}{"\n"}'
kubectl get fi <fi-name> -o jsonpath='{.status.last_build.job_ref.name}{"\n"}'
kubectl get fi <fi-name> -o jsonpath='{.status.bundle_ref.name}{"\n"}'
```

## Update

Prefer editing YAML and re-applying it:

```bash
kubectl apply -f <fi.yaml>
```

Use patch only for explicit targeted changes.

## Disable

```bash
kubectl patch fi <fi-name> --type=merge -p '{"spec":{"enabled":false}}'
kubectl get fi <fi-name> -o yaml
```

## Enable

```bash
kubectl patch fi <fi-name> --type=merge -p '{"spec":{"enabled":true}}'
kubectl get fi <fi-name> -o yaml
```

## Delete

```bash
kubectl delete fi <fi-name>
```

## Key Fields To Inspect

- `.status.phase`
- `.status.message`
- `.status.observed_generation`
- `.status.observed_spec_hash`
- `.status.last_build.job_ref.name`
- `.status.last_build.started_at`
- `.status.bundle_ref.name`
