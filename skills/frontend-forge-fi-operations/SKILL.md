---
name: frontend-forge-fi-operations
description: Operate FrontendIntegration resources and the frontend-forge extension. Use when Codex needs to create a FrontendIntegration from FrontendIntegration YAML, update or patch FI lifecycle state, inspect or troubleshoot FI build output, or create, enable, disable, uninstall, and inspect the frontend-forge extension through its InstallPlan and extension resources.
---

# Frontend Forge FI Operations

Operate `FrontendIntegration` resources and the `frontend-forge` extension with `kubectl`.

## When to Use

- Create a `FrontendIntegration` from complete `FrontendIntegration` YAML
- Update, enable, disable, or delete an existing FI
- Inspect FI status, build jobs, `JSBundle`, ConfigMap, manifest, or source-spec annotations
- Troubleshoot FI reconciliation or build output
- Create, enable, disable, uninstall, or inspect the `frontend-forge` extension

## Do Not Use

- Generate new `FrontendIntegration` YAML from natural language
- Manage extensions other than `frontend-forge`
- Assume FI operations are valid before checking extension state
- Use patch as the creation path for a new FI when the user already has YAML

## Preconditions

- `kubectl` must be configured for the target cluster.
- `frontend-forge` extension installation and enabled state are prerequisites for all FI functionality.
- `frontend-forge-controller` exists only after the `frontend-forge` extension is installed.
- `FrontendIntegration` is cluster-scoped. Do not add `-n` when operating on FI resources.

## Read First

- Read [references/extension-management.md](references/extension-management.md) before any FI operation.
- Read [references/lifecycle.md](references/lifecycle.md) for FI create, update, enable, disable, and delete workflows.
- Read [references/inspection.md](references/inspection.md) for FI inspection and troubleshooting.

## Preflight

Always run these checks before any FI create, update, enable, disable, delete, inspect, or troubleshoot flow:

```bash
kubectl get extension frontend-forge
kubectl get installplan frontend-forge -o yaml
```

Only continue to FI operations when:

- the `frontend-forge` extension exists
- the `frontend-forge` InstallPlan exists
- `spec.enabled=true` on the InstallPlan

If the extension is missing or disabled, switch to the extension management workflow first.

## FI Operations

### Create FI From YAML

Only create FI from complete `FrontendIntegration` YAML content or a YAML file path.

1. Run the preflight checks.
2. If the user provides YAML content, write it to a temporary file.
3. Apply the YAML with `kubectl apply -f`.
4. Inspect the result.

See [references/lifecycle.md](references/lifecycle.md) for the exact create and post-apply inspection commands.

### Update Existing FI

- Prefer editing YAML and re-applying it.
- Use patch only for targeted lifecycle changes when the user explicitly wants patch semantics.

See [references/lifecycle.md](references/lifecycle.md) for the exact update, enable, disable, and delete commands.

### Disable FI

Disable an existing FI by patching `spec.enabled=false`.

See [references/lifecycle.md](references/lifecycle.md) for the exact disable command and post-check flow.

### Enable FI

Enable an existing FI by patching `spec.enabled=true`.

See [references/lifecycle.md](references/lifecycle.md) for the exact enable command and post-check flow.

### Delete FI

Delete the FI resource only after the preflight checks pass.

See [references/lifecycle.md](references/lifecycle.md) for the exact delete command.

## Extension Management

Use fixed resource names for `frontend-forge`:

- extension: `frontend-forge`
- InstallPlan: `frontend-forge`

### Create Extension

Ask the user for the `frontend-forge` extension version. If it is not provided, resolve the latest version first.

See [references/extension-management.md](references/extension-management.md) for the exact version lookup command and InstallPlan YAML.

### Disable Extension

Patch `installplan/frontend-forge` so `spec.enabled=false`.

### Enable Extension

Patch `installplan/frontend-forge` so `spec.enabled=true`.

### Uninstall Extension

Delete `installplan/frontend-forge`.

### Inspect Extension State

Inspect `extension/frontend-forge` and `installplan/frontend-forge`.

See [references/extension-management.md](references/extension-management.md) for the exact create, enable, disable, uninstall, and inspect commands.

## Rules

- Treat extension readiness as the first gate for all FI operations.
- Create new FI resources only from `FrontendIntegration` YAML.
- Prefer `kubectl apply -f` for FI creation.
- Keep FI troubleshooting available after the preflight checks pass.
- Keep command details in references; keep `SKILL.md` focused on routing and decision logic.
- Use `extension` and `installplan` resources as the source of truth for extension state.
- Treat `frontend-forge-controller` existence as a runtime signal, not the primary source of truth.
- Do not generalize extension commands to any name other than `frontend-forge`.
