# FrontendExtension Inspection And Troubleshooting Reference

Primary sources:

- FE object status and labels
- package, publish, and unpublish Jobs
- artifact ConfigMaps
- publish target ConfigMap or Secret
- FE controller and FE API logs

## Inspect The FE First

```bash
kubectl get fe <name> -o yaml
kubectl get fe <name> -o jsonpath='{.status.phase}{"\n"}'
kubectl get fe <name> -o jsonpath='{.status.conditions}{"\n"}'
```

Important fields:

- `.status.phase`
- `.status.observedGeneration`
- `.status.observedSourceHash`
- `.status.observedRebuildToken`
- `.status.conditions`
- `.status.packageJob`
- `.status.artifact`
- `.status.download`
- `.status.publish`
- `.status.unpublish`

Implementation-backed status values:

- `.status.phase`: `Pending`, `Packaging`, `Ready`, `Failed`
- `.status.packageJob.phase`: `Pending`, `Running`, `Succeeded`, `Failed`
- `.status.publish.phase`: `NotRequested`, `Pending`, `Running`, `Succeeded`, `Failed`
- `.status.unpublish.phase`: `NotRequested`, `Pending`, `Running`, `Succeeded`, `Failed`
- condition types: `SourceValid`, `ArtifactReady`, `DownloadReady`, `PublishSucceeded`
- condition reasons include `Validated`, `Packaging`, `Generated`, `ArtifactNotReady`, `Available`, `NotRequested`, `Succeeded`, `PublishFailed`, `InvalidSource`, and package failure reasons such as `PackageAttemptsExceeded`

## Package Job

Package Job reference from status:

```bash
kubectl get fe <name> -o jsonpath='{.status.packageJob.namespace}{" "}{.status.packageJob.name}{"\n"}'
```

List all related Jobs:

```bash
kubectl get jobs -n extension-frontend-forge -l frontend-forge.kubesphere.io/fe-name=<name>
```

Inspect logs:

```bash
kubectl -n <job-namespace> describe job <job-name>
kubectl -n <job-namespace> logs job/<job-name>
```

Common package failures:

- invalid inline source or menu/page binding
- build-service endpoint unreachable
- stale source hash detected by the package Job
- missing selected bundle key from build-service result
- artifact ConfigMap missing or metadata mismatched
- package attempts exceeded

## Artifact ConfigMap

Find artifact storage from FE status:

```bash
kubectl get fe <name> -o jsonpath='{.status.artifact.storage.ref.namespace}{" "}{.status.artifact.storage.ref.name}{" "}{.status.artifact.storage.key}{"\n"}'
```

Inspect metadata:

```bash
kubectl -n <artifact-namespace> get cm <artifact-configmap-name> -o yaml
```

Expected ConfigMap data:

- `binaryData["package.tgz"]`
- `data["artifact.json"]`
- `data["files.json"]`

Expected annotations:

- `frontend-forge.kubesphere.io/source-hash`
- `frontend-forge.kubesphere.io/artifact-key`
- `frontend-forge.kubesphere.io/artifact-digest`
- `frontend-forge.kubesphere.io/artifact-filename`

The artifact ConfigMap source hash and artifact key must match FE status. The package digest is `status.artifact.digest`; the artifact key is cache identity.

Safety rules:

- Treat the artifact ConfigMap as controller-owned output.
- Prefer `GET /frontendextensions/<name>/download` for package bytes because the API verifies readiness, storage kind, source hash, and digest.
- Do not edit `binaryData["package.tgz"]`, `data["artifact.json"]`, or `data["files.json"]`.
- Do not delete the ConfigMap referenced by `.status.artifact.storage.ref` unless the user requested cleanup and you confirmed it is not the current artifact or the FE is being deleted.

## Publish State

Inspect publish status:

```bash
kubectl get fe <name> -o jsonpath='{.status.publish}{"\n"}'
```

Prefer the FE API for publish:

```bash
KS_API=https://<kubesphere-host>
FE_API="$KS_API/kapis/frontend-forge-api.kubesphere.io/v1alpha1/frontendextensions"
curl -fS "$FE_API/<name>/publish"
curl -fS -u "user:password" "$FE_API/<name>/publish"
digest=$(kubectl get fe <name> -o jsonpath='{.status.artifact.digest}')
curl -fS -X POST -H 'Content-Type: application/json' --data "{\"requestId\":\"manual-1\",\"expectedArtifactDigest\":\"${digest}\"}" "$FE_API/<name>/publish"
```

`GET "$FE_API/<name>/publish"` is read-only; use it to inspect current publish status before or after the POST.
If `/kapis` returns `401` or `403`, show the `curl -u "user:password"` form as a user-run command and treat the issue as KubeSphere authentication or authorization.

Annotations are diagnostic evidence normally written by the API:

- `frontend-forge.kubesphere.io/publish-request-id`
- `frontend-forge.kubesphere.io/publish-request-generation`
- `frontend-forge.kubesphere.io/publish-request-source-hash`
- `frontend-forge.kubesphere.io/publish-artifact-digest`
- `frontend-forge.kubesphere.io/publish-target-kind`
- `frontend-forge.kubesphere.io/publish-target-namespace`
- `frontend-forge.kubesphere.io/publish-target-name`

Publisher Job:

```bash
kubectl get fe <name> -o jsonpath='{.status.publish.jobRef.namespace}{" "}{.status.publish.jobRef.name}{"\n"}'
kubectl -n <job-namespace> describe job <job-name>
kubectl -n <job-namespace> logs job/<job-name>
```

Check publish target:

```bash
kubectl -n <target-namespace> get cm <target-name> -o yaml
kubectl -n <target-namespace> get secret <target-name> -o yaml
```

Publish target kind must be `ConfigMap` or `Secret`. Target data is passed to `ksbuilder publish`; keys `env.<NAME>` become environment variables and key `args` is split into additional args.

## Unpublish State

Inspect unpublish status:

```bash
kubectl get fe <name> -o jsonpath='{.status.unpublish}{"\n"}'
```

Prefer the FE API for unpublish:

```bash
KS_API=https://<kubesphere-host>
FE_API="$KS_API/kapis/frontend-forge-api.kubesphere.io/v1alpha1/frontendextensions"
curl -fS "$FE_API/<name>/unpublish"
curl -fS -u "user:password" "$FE_API/<name>/unpublish"
curl -fS -X POST -H 'Content-Type: application/json' --data '{"requestId":"manual-unpublish-1"}' "$FE_API/<name>/unpublish"
```

`GET "$FE_API/<name>/unpublish"` is read-only; use it to inspect current unpublish status before or after the POST.
If `/kapis` returns `401` or `403`, show the `curl -u "user:password"` form as a user-run command and treat the issue as KubeSphere authentication or authorization.

Annotations are diagnostic evidence normally written by the API:

- `frontend-forge.kubesphere.io/unpublish-request-id`
- `frontend-forge.kubesphere.io/unpublish-extension-name`
- `frontend-forge.kubesphere.io/delete-after-unpublish-request-id`
- publish target annotations

Unpublish Job:

```bash
kubectl get fe <name> -o jsonpath='{.status.unpublish.jobRef.namespace}{" "}{.status.unpublish.jobRef.name}{"\n"}'
kubectl -n <job-namespace> describe job <job-name>
kubectl -n <job-namespace> logs job/<job-name>
```

## Status Labels

Use labels for list filtering and high-level triage:

```bash
kubectl get fe -l frontend-forge.kubesphere.io/package-state=packaging
kubectl get fe -l frontend-forge.kubesphere.io/package-state=failed
kubectl get fe -l frontend-forge.kubesphere.io/publish-state=published
kubectl get fe -l frontend-forge.kubesphere.io/publish-fresh=false
```

Label meanings:

- `package-state=packaging`: pending or packaging package artifact
- `package-state=ready`: current artifact is available
- `package-state=failed`: source validation or package generation failed
- `publish-state=not-published`: no active succeeded publish for current state
- `publish-state=publishing`: publish or unpublish work is pending/running
- `publish-state=published`: active succeeded publish
- `publish-fresh=false`: active publish digest does not match current ready artifact digest, or no fresh publish exists

## Controller Configuration

Inspect env vars when defaults do not match live behavior:

```bash
kubectl -n extension-frontend-forge get deploy -l app.kubernetes.io/component=extension-controller -o yaml
kubectl -n extension-frontend-forge get deploy -l app.kubernetes.io/component=extension-api -o yaml
```

Important FE controller env vars:

- `WORK_NAMESPACE`
- `PACKAGER_IMAGE`
- `PACKAGER_SERVICE_ACCOUNT`
- `PUBLISHER_IMAGE`
- `PUBLISHER_SERVICE_ACCOUNT`
- `ARTIFACT_CONFIGMAP_NAMESPACE`
- `BUILD_SERVICE_BASE_URL`
- `BUILD_SERVICE_TIMEOUT_SECONDS`
- `JSBUNDLE_CONFIG_KEY`
- `RECONCILE_REQUEUE_SECONDS`
- `JOB_ACTIVE_DEADLINE_SECONDS`
- `JOB_TTL_SECONDS_AFTER_FINISHED`
- `ARTIFACT_RETAIN_OLD_COUNT`
- `PACKAGE_MAX_ATTEMPTS`

## Debugging Patterns

Package stuck in `Packaging`:

1. Check package Job phase and logs.
2. Check build-service reachability from the Job.
3. Check controller logs for requeue or create errors.
4. Check whether the package Job is active until the deadline.

Package `Failed` without a useful artifact:

1. Check `SourceValid` and `ArtifactReady` conditions.
2. Inspect package Job message and logs.
3. Confirm `PACKAGE_MAX_ATTEMPTS`.
4. Check whether the artifact ConfigMap exists but has mismatched annotations.

Ready but download fails:

1. Confirm `status.download.ready=true`.
2. Confirm `status.artifact.storage.kind=ConfigMap`.
3. Confirm the ConfigMap and key from status exist.
4. Confirm digest in status matches artifact bytes if downloaded manually.

Publish failed:

1. Confirm package artifact is ready or publish was allowed to wait for artifact.
2. Check target kind/ref from `spec.publishPolicy`, API response, and publish annotations.
3. Inspect publisher Job logs.
4. Confirm `ksbuilder` target data and credentials in the target ConfigMap or Secret.

Published but stale:

1. Compare `status.publish.artifactDigest` with `status.artifact.digest`.
2. Check `frontend-forge.kubesphere.io/publish-fresh`.
3. Trigger a new publish request through the FE API.
