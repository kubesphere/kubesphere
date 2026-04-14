# FI Runtime Inspection Reference

Extension must be installed and enabled before FI troubleshooting. See `SKILL.md` Preflight.

Remember: `frontend-forge-controller` exists only after the extension is installed.

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

## Important Labels And Annotations

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

## Troubleshooting Order

1. Check FI `status`
2. Check `.status.bundle_ref.name` and `.status.last_build.job_ref.name`
3. Inspect Job logs, JSBundle annotations, and ConfigMap content
4. If the controller appears missing, verify the `frontend-forge` extension and InstallPlan state first
