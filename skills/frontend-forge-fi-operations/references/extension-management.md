# Frontend Forge Extension Management Reference

## Resource Model

- extension name: `frontend-forge`
- InstallPlan name: `frontend-forge`

Use these resources as the source of truth:

```bash
kubectl get extension frontend-forge
kubectl get installplan frontend-forge -o yaml
```

## Resolve Latest Version

If the user does not provide a version, query the latest available version:

```bash
kubectl get extensionversions -n kubesphere-system -l kubesphere.io/extension-ref=frontend-forge -o jsonpath='{range .items[*]}{.spec.version}{"\n"}{end}' | sort -V | tail -1
```

## Create

Apply this InstallPlan:

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: frontend-forge
spec:
  enabled: true
  extension:
    name: frontend-forge
    version: <version>
```

`InstallPlan` is cluster-scoped. Do not add `metadata.namespace`.

## Disable

```bash
kubectl patch installplan frontend-forge --type=merge -p '{"spec":{"enabled":false}}'
```

## Enable

```bash
kubectl patch installplan frontend-forge --type=merge -p '{"spec":{"enabled":true}}'
```

## Uninstall

```bash
kubectl delete installplan frontend-forge
```

## Inspect State

```bash
kubectl get extension frontend-forge
kubectl get installplan frontend-forge -o yaml
```

## Notes

- `frontend-forge-controller` is created only after the extension is installed.
- For FI operations, extension installation and enabled state are mandatory prerequisites.
