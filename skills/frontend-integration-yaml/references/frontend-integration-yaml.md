# Frontend Integration YAML Reference

## Overview

This skill generates canonical Kubernetes resources for:

```yaml
apiVersion: frontend-forge.kubesphere.io/v1alpha1
kind: FrontendIntegration
```

The authoring input is intentionally simpler than the output. Write a single-menu JSON spec, then let the generator expand it into canonical `spec.menus[]`.

## Input Schema

The generator accepts JSON from stdin or from `--input <path>`.

### Top-Level Fields

```json
{
  "metadata": {
    "name": "required",
    "annotations": {
      "kubesphere.io/description": "optional"
    }
  },
  "spec": {
    "enabled": true,
    "displayName": "optional",
    "builder": {
      "engineVersion": "optional"
    },
    "locales": {
      "en": {
        "KEY": "Value"
      }
    }
  },
  "menu": {
    "displayName": "required",
    "icon": "optional, defaults to GridDuotone",
    "placements": ["cluster"]
  },
  "pages": []
}
```

### Page Variants

#### `crdTable`

```json
{
  "displayName": "Bundles",
  "key": "optional",
  "type": "crdTable",
  "crdTable": {
    "authKey": "optional",
    "group": "extensions.kubesphere.io",
    "version": "v1alpha1",
    "scope": "Cluster",
    "names": {
      "kind": "JSBundle",
      "plural": "jsbundles"
    },
    "columns": []
  }
}
```

Rules:

- `group`, `version`, `scope`, and `names.plural` are required.
- `names.kind` is optional and passed through when provided.
- `key` defaults to `names.plural`.
- `columns` is optional. When omitted, default columns are derived from scope.

#### `iframe`

```json
{
  "displayName": "Dashboard",
  "key": "dashboard",
  "type": "iframe",
  "iframe": {
    "src": "https://example.test/dashboard"
  }
}
```

Rules:

- `key` is required.
- `iframe.src` is required.

## Output Shape

The generator always outputs canonical `FrontendIntegration` YAML:

```yaml
apiVersion: frontend-forge.kubesphere.io/v1alpha1
kind: FrontendIntegration
metadata:
  name: example
spec:
  enabled: true
  menus:
    - key: example-cluster
      displayName: Operations
      icon: BoxDuotone
      placement: cluster
      type: organization
      children:
        - key: jsbundles
          displayName: Bundles
  pages:
    - key: jsbundles
      type: crdTable
      crdTable:
        group: extensions.kubesphere.io
        version: v1alpha1
        scope: Cluster
        names:
          kind: JSBundle
          plural: jsbundles
        columns:
          - key: name
            title: NAME
            enableSorting: true
            render:
              type: text
              path: metadata.name
          - key: updateTime
            title: CREATION_TIME
            enableHiding: true
            enableSorting: true
            render:
              type: time
              path: metadata.creationTimestamp
              format: local-datetime
```

## Field Mapping

- `metadata.name` maps directly to the resource name.
- `metadata.annotations` is preserved when provided.
- `spec.enabled` defaults to `true`.
- `menu.icon` defaults to `GridDuotone`.
- `menu.displayName`, `menu.icon`, and `menu.placements[]` expand into `spec.menus[]`.
- Each canonical menu key is `${metadata.name}-${placement}`.
- Each menu uses `type: organization`.
- `pages[].displayName` becomes the corresponding menu child `displayName`.
- `pages[].key` becomes both the page key and the child key.

## Normalization Rules

### Placement Rules

- Supported placements: `cluster`, `workspace`, `global`.
- Duplicate placements are rejected.
- If any placement is `workspace`, every `crdTable.scope` is forced to `Namespaced`.

### Scope Rules

- Accept `Cluster`, `cluster`, `Namespaced`, `namespaced`, `Namespace`, or `namespace`.
- Normalize accepted namespaced values to `Namespaced`.
- Normalize cluster values to `Cluster`.

### Default Columns

For `Cluster`:

```yaml
columns:
  - enableSorting: true
    key: name
    render:
      path: metadata.name
      type: text
    title: NAME
  - enableHiding: true
    enableSorting: true
    key: updateTime
    render:
      format: local-datetime
      path: metadata.creationTimestamp
      type: time
    title: CREATION_TIME
```

For `Namespaced`:

```yaml
columns:
  - enableSorting: true
    key: name
    render:
      path: metadata.name
      type: text
    title: NAME
  - enableHiding: true
    key: namespace
    render:
      path: metadata.namespace
      type: text
    title: PROJECT
  - enableHiding: true
    enableSorting: true
    key: updateTime
    render:
      format: local-datetime
      path: metadata.creationTimestamp
      type: time
    title: CREATION_TIME
```

### Provided Column Normalization

- Remove any existing `namespace` column when the final scope is `Cluster`.
- Ensure exactly one `namespace` column when the final scope is `Namespaced`.
- Insert `namespace` before `updateTime` when `updateTime` exists.
- Otherwise insert `namespace` after `name` when `name` exists.
- Otherwise append `namespace` at the end.

## Validation Rules

The generator rejects:

- missing `metadata.name`
- missing `menu.displayName` or `menu.placements`
- duplicate placements
- duplicate page keys
- missing required `crdTable` fields
- missing `iframe.src`

## Examples

### CRD Table Only

Input:

```json
{
  "metadata": {
    "name": "bundles"
  },
  "menu": {
    "displayName": "Extensions",
    "icon": "BoxDuotone",
    "placements": ["cluster"]
  },
  "pages": [
    {
      "displayName": "Bundles",
      "type": "crdTable",
      "crdTable": {
        "group": "extensions.kubesphere.io",
        "version": "v1alpha1",
        "scope": "Cluster",
        "names": {
          "kind": "JSBundle",
          "plural": "jsbundles"
        }
      }
    }
  ]
}
```

Result:

- outputs canonical `FrontendIntegration`
- creates `menus[0].key = bundles-cluster`
- derives page key `jsbundles`
- generates the cluster default columns

### Mixed CRD And Iframe

Input:

```json
{
  "metadata": {
    "name": "ops"
  },
  "menu": {
    "displayName": "Operations",
    "icon": "AppsGearDuotone",
    "placements": ["cluster", "workspace"]
  },
  "pages": [
    {
      "displayName": "Bundles",
      "type": "crdTable",
      "crdTable": {
        "group": "extensions.kubesphere.io",
        "version": "v1alpha1",
        "scope": "Cluster",
        "names": {
          "plural": "jsbundles"
        }
      }
    },
    {
      "displayName": "Dashboard",
      "key": "dashboard",
      "type": "iframe",
      "iframe": {
        "src": "https://example.test/dashboard"
      }
    }
  ]
}
```

Result:

- expands to `ops-cluster` and `ops-workspace`
- forces the CRD page scope to `Namespaced`
- inserts the `PROJECT` / `namespace` column
