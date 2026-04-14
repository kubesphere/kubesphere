---
name: frontend-integration-yaml
description: Generate canonical FrontendIntegration YAML from a simplified single-menu authoring model for frontend-forge.
---

# Frontend Integration YAML

Generate submit-ready `FrontendIntegration` YAML from a simplified authoring model with one menu and one or more pages.

## When to Use

- generate FrontendIntegration YAML
- create a quick integration YAML from a CRD
- create an iframe based FrontendIntegration YAML
- convert simplified single-menu config into canonical YAML

## Do Not Use

- troubleshooting existing FrontendIntegration resources
- managing enable/disable/delete lifecycle
- inspecting JSBundle, Job, controller, or status flows
- editing manifest generation rules
- handling multi-menu authoring models

## Required Inputs

Ask the user instead of guessing when any of these are missing:
- `metadata.name`
- `menu.displayName`
- `menu.placements[]`
- at least one page in `pages[]`
- for `crdTable`: `group`, `version`, `scope`, `names.plural`
- for `iframe`: `key` and `iframe.src`

## Authoring Model

Use the simplified single-menu model as input to the generator:
- top-level `metadata`
- top-level `spec` for optional passthrough fields
- top-level `menu`
- top-level `pages`

Do not author canonical `spec.menus[]` by hand.

## Workflow

1. Extract the simplified authoring model.
2. Fill only safe defaults.
3. Keep the model in single-menu form.
4. Run `scripts/generate_frontend_integration.py`.
5. Return the generated canonical YAML.

## Rules

- Treat the generator script as the source of truth.
- Do not manually reconstruct canonical YAML when the generator is available.
- Always output canonical `kind: FrontendIntegration`.
- Always expand `menu.placements[]` into `spec.menus[]`.
- Always derive `children` from `pages[]`.
- If any placement is `workspace`, force all `crdTable.scope` values to `Namespaced`.
- Default `menu.icon` to `GridDuotone` when it is omitted.
- Do not add runtime-managed metadata unless explicitly requested.

## Output

- Return the generated YAML first, unless explanation is requested.
- Prefer a single `yaml` code block.
- After the YAML, always output this exact guidance line:
  `FrontendIntegration YAML has been generated. Reply 'apply' to deploy it directly.`
