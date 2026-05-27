---
name: frontend-forge-fe-operations
description: "Operate FrontendExtension (FE) resources in frontend-forge: create, update, rebuild, inspect package artifacts, download packages, publish, unpublish, delete, and debug package/publish controller behavior. Use this skill whenever the user mentions FE operations, FrontendExtension lifecycle, extension package/download/publish/unpublish, artifact ConfigMaps, package Jobs, publisher Jobs, publish target ConfigMaps or Secrets, rebuild-token, package-state/publish-state labels, or troubleshooting FE status in a Kubernetes cluster."
---

# Frontend Forge FE Operations

## When to use

Use this skill for operational work around `FrontendExtension` resources:

- Create or update an FE manifest
- Trigger or inspect package generation
- Force a rebuild with the rebuild-token annotation
- Download the generated package artifact
- Publish, unpublish, or delete an extension
- Inspect FE status, conditions, labels, Jobs, artifact ConfigMaps, and publish targets
- Debug FE package phases such as `Packaging`, `Ready`, and `Failed`, plus publish/unpublish phases such as `Pending`, `Running`, `Succeeded`, and `Failed`

If the task is about `FrontendIntegration` runtime `JSBundle` creation, use the FI operations skill instead. FE package/publish does not create a runtime `JSBundle` in the current cluster.

## Quick Entry

- Create or update FE -> read `references/lifecycle.md`
- Rebuild package -> read `references/lifecycle.md`
- Publish, unpublish, delete-with-unpublish, list through API, or download -> read `references/api.md`
- Inspect package Job, artifact ConfigMap, publish Job, labels, or conditions -> read `references/inspection.md`
- Debug stuck or failed state -> read `references/inspection.md`, then check controller logs

## Preconditions

- The current Kubernetes context is pointed at the target cluster.
- The FE CRD exists: `frontendextensions.frontend-forge.kubesphere.io`.
- The FE controller is installed, usually as the `extension-controller` component.
- The FE API is installed when using download, publish, unpublish, or delete-with-unpublish HTTP operations.
- The build service configured by `BUILD_SERVICE_BASE_URL` is reachable from package Jobs.
- Publish target data exists when publishing, normally `ConfigMap/ksbuilder-publish-config` in `extension-frontend-forge`.

## Source of truth

Prefer live cluster state for operations and repo docs for expected behavior:

- `kubectl get fe <name> -o yaml`
- `kubectl get jobs -n <work-namespace> -l frontend-forge.kubesphere.io/fe-name=<name>`
- `spec/frontend-extension-design.md`
- `spec/crds.md`
- `spec/k8s-resources.md`
- `config/samples/frontendextension-inspecttask.yaml`
- `crates/api/src/fe.rs`
- `crates/frontend-extension-controller`
- `crates/frontend-forge-extension-api`

## Resource Model

- `FrontendExtension`
  - cluster-scoped
  - short name: `fe`
  - source object for package, artifact, download, publish, and unpublish state
- Package Job
  - namespaced
  - default namespace: `extension-frontend-forge`
  - typical name: `fe-<fe-name>-package-<artifact-key-12>-a<attempt>`
- Artifact ConfigMap
  - namespaced
  - default namespace: `extension-frontend-forge`
  - referenced by `status.artifact.storage.ref`
  - contains `binaryData["package.tgz"]`, `data["artifact.json"]`, and `data["files.json"]`
- Publish or unpublish Job
  - namespaced
  - default namespace: `extension-frontend-forge`
  - typical names:
    - `fe-<fe-name>-publish-<request-id-hash-short>`
    - `fe-<fe-name>-unpublish-<request-id-hash-short>`
- Publish target
  - `ConfigMap` or `Secret`
  - default chart target is usually `ConfigMap/ksbuilder-publish-config` in the release namespace

Default names and namespaces can change through Helm values and controller environment variables. Use FE status, Job labels, and controller deployment env vars before assuming defaults.

## Common Commands

Inspect FE:

```bash
kubectl get fe <name> -o yaml
kubectl get fe <name> -o jsonpath='{.status}{"\n"}'
kubectl get fe <name> -o jsonpath='{.status.phase}{" "}{.status.publish.phase}{" "}{.status.unpublish.phase}{"\n"}'
kubectl get fe -l frontend-forge.kubesphere.io/package-state=ready
kubectl get fe -l frontend-forge.kubesphere.io/publish-state=published
```

Create or update:

```bash
kubectl apply -f <file.yaml>
kubectl apply -f config/samples/frontendextension-inspecttask.yaml
```

Find package Job and logs:

```bash
kubectl get fe <name> -o jsonpath='{.status.packageJob.namespace}{" "}{.status.packageJob.name}{"\n"}'
kubectl -n <job-namespace> get job <job-name> -o yaml
kubectl -n <job-namespace> logs job/<job-name>
```

Find artifact ConfigMap:

```bash
kubectl get fe <name> -o jsonpath='{.status.artifact.storage.ref.namespace}{" "}{.status.artifact.storage.ref.name}{" "}{.status.artifact.storage.key}{"\n"}'
kubectl -n <artifact-namespace> get cm <artifact-configmap-name> -o yaml
```

Force a rebuild:

```bash
kubectl annotate fe <name> frontend-forge.kubesphere.io/rebuild-token="$(date +%s)" --overwrite
```

Inspect publish or unpublish Jobs:

```bash
kubectl get jobs -n extension-frontend-forge -l frontend-forge.kubesphere.io/fe-name=<name>
kubectl get fe <name> -o jsonpath='{.status.publish.jobRef.namespace}{" "}{.status.publish.jobRef.name}{"\n"}'
kubectl get fe <name> -o jsonpath='{.status.unpublish.jobRef.namespace}{" "}{.status.unpublish.jobRef.name}{"\n"}'
kubectl -n <job-namespace> logs job/<job-name>
```

FE API operations:

```bash
KS_API=https://<kubesphere-host>
FE_API="$KS_API/kapis/frontend-forge-api.kubesphere.io/v1alpha1/frontendextensions"
curl -fS "$FE_API/<name>"
curl -fS -u "user:password" "$FE_API/<name>"  # user runs this if /kapis returns 401/403
curl -fS "$FE_API/<name>/publish"      # read publish status only
curl -fS -X POST -H 'Content-Type: application/json' --data '{"requestId":"manual-1","expectedArtifactDigest":"sha256:<digest>"}' "$FE_API/<name>/publish"
curl -fS "$FE_API/<name>/unpublish"    # read unpublish status only
curl -fS -X POST -H 'Content-Type: application/json' --data '{"requestId":"manual-unpublish-1"}' "$FE_API/<name>/unpublish"
curl -fS -X POST -H 'Content-Type: application/json' --data '{"unpublish":true}' "$FE_API/<name>/delete"
curl -fS -u "user:password" -X POST -H 'Content-Type: application/json' --data '{"unpublish":true}' "$FE_API/<name>/delete"
curl -fL "$FE_API/<name>/download" -o <name>.tgz
```

## Operating Rules

- Do not use `-n` with `kubectl get fe`; `FrontendExtension` is cluster-scoped.
- Use the FE HTTP API for publish, unpublish, download, and delete-with-unpublish when available. It validates artifact readiness, publish target, digest expectations, and idempotency.
- For KubeSphere users, prefer `/kapis/frontend-forge-api.kubesphere.io/v1alpha1/frontendextensions`; use direct service port-forwarding mainly for local debugging.
- If `/kapis` returns `401` or `403`, guide the user to run `curl -u "user:password" ...` or use their normal KubeSphere authenticated session. Do not ask them to paste credentials into the conversation.
- Treat publish and unpublish annotations as implementation details written by the FE API. Inspect them for debugging, but do not make raw annotation patches the default workflow.
- Directly patch publish/unpublish annotations only as a last-resort recovery step when the API is unavailable and the operator explicitly accepts the risk. Missing or stale generation, source hash, artifact digest, or target annotations can produce failed publish status.
- Do not expect package creation to publish the extension. Packaging and publishing are separate flows.
- Do not expect direct `kubectl delete fe <name>` to unpublish first. Use the FE API delete endpoint with `{"unpublish":true}` for that behavior.
- Treat `status.artifact.artifactKey` as build cache identity, not package content digest. Use `status.artifact.digest` for the generated package digest.
- Treat artifact ConfigMaps as controller-owned outputs. Read them for inspection, but do not edit or delete the ConfigMap referenced by `status.artifact.storage.ref` unless the user has explicitly requested artifact cleanup and you have confirmed it is not the current artifact.

## Safety Rules

Destructive operations:

- Before deleting an FE, check `.status.publish.phase`, `.status.publish.active`, and `.status.publish.artifactDigest`.
- If a published extension should be removed, use the FE API delete endpoint with `{"unpublish":true}` so the controller can unpublish before deleting.
- Avoid `kubectl delete fe <name>` for published extensions unless the user explicitly wants to skip unpublish.
- Before deleting Jobs or artifact ConfigMaps, verify owner refs, labels, and whether FE status still references them.

Rebuild:

- Rebuild changes artifact cache identity and can make a previously published artifact stale. Before setting `frontend-forge.kubesphere.io/rebuild-token`, record current `.status.observedSourceHash`, `.status.artifact.digest`, `.status.artifact.artifactKey`, `.status.publish`, and `frontend-forge.kubesphere.io/publish-fresh`.
- Prefer a unique rebuild token, such as a timestamp or incident id, and use `--overwrite`.
- After rebuild, verify the new package reaches `Ready`; republish only after confirming the new `status.artifact.digest`.

Artifact ConfigMap:

- Use the FE API download endpoint for package bytes when possible; it checks phase, download readiness, source hash, storage kind, and digest.
- Do not mutate `binaryData["package.tgz"]`, `data["artifact.json"]`, or `data["files.json"]` in place. A manual edit can make status, annotations, and digest disagree.
- Do not delete the ConfigMap named by `.status.artifact.storage.ref` during normal troubleshooting. If cleanup is requested, first confirm the current FE artifact points somewhere else or the FE itself is being deleted.

## Status Fields To Check

FE package state:

- `.status.phase`
- `.status.observedGeneration`
- `.status.observedSourceHash`
- `.status.observedRebuildToken`
- `.status.conditions`
- `.status.packageJob`
- `.status.artifact`
- `.status.download`

Publish state:

- `.status.publish.phase`
- `.status.publish.active`
- `.status.publish.requestId`
- `.status.publish.artifactDigest`
- `.status.publish.jobRef`
- `.status.publish.lastError`
- `.status.unpublish.phase`
- `.status.unpublish.requestId`
- `.status.unpublish.extensionName`
- `.status.unpublish.jobRef`
- `.status.unpublish.lastError`

Labels used for filtering:

- `frontend-forge.kubesphere.io/package-state`: `packaging`, `ready`, `failed`
- `frontend-forge.kubesphere.io/publish-state`: `not-published`, `publishing`, `published`, `failed`
- `frontend-forge.kubesphere.io/publish-fresh`: `true`, `false`

Status consistency with implementation:

- FE package phase is only `Pending`, `Packaging`, `Ready`, or `Failed`.
- Package Job phase is `Pending`, `Running`, `Succeeded`, or `Failed`.
- Publish phase is `NotRequested`, `Pending`, `Running`, `Succeeded`, or `Failed`.
- Unpublish phase is `NotRequested`, `Pending`, `Running`, `Succeeded`, or `Failed`.
- Controller condition types are `SourceValid`, `ArtifactReady`, `DownloadReady`, and `PublishSucceeded`.
- `PublishFailed` appears as a `PublishSucceeded` condition reason when `status.publish.phase=Failed`; it is not a top-level FE phase.

## Troubleshooting Order

1. Inspect the FE object and status first.
2. If package state is not `Ready`, inspect `status.conditions`, `status.packageJob`, package Job logs, build-service reachability, and artifact ConfigMap state.
3. If package state is `Ready` but download fails, inspect `status.download`, `status.artifact.storage`, the artifact ConfigMap key, and digest consistency.
4. If publish or unpublish fails, inspect `status.publish` or `status.unpublish`, publisher Job logs, target ConfigMap or Secret contents, and API responses. Use annotations as diagnostic evidence, not as the primary operation path.
5. If live behavior differs from expected defaults, inspect controller deployment env vars and Helm values before changing the FE.

## Controller Logs

When status and Job logs are insufficient:

```bash
kubectl -n extension-frontend-forge logs deploy/frontend-forge-extension-controller --tail=200
kubectl -n extension-frontend-forge logs deploy/frontend-forge-extension-api --tail=200
```

Deployment names can vary by Helm release name. If those commands fail, list deployments with:

```bash
kubectl -n extension-frontend-forge get deploy -l app.kubernetes.io/component=extension-controller
kubectl -n extension-frontend-forge get deploy -l app.kubernetes.io/component=extension-api
```
