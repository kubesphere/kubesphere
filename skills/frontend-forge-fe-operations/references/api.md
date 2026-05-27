# FrontendExtension API Operations Reference

Use the FE HTTP API for list/get/create/download/publish/unpublish/delete operations. For KubeSphere users, prefer the aggregated `/kapis/frontend-forge-api.kubesphere.io/v1alpha1/...` path. It applies validation that raw annotation patches skip, so it should be the default operation path.

## API Base Paths

Prefer the KubeSphere aggregated route:

- `/kapis/frontend-forge-api.kubesphere.io/v1alpha1/frontendextensions`

The service also serves compatibility routes:

- `/apis/frontend-forge.kubesphere.io/v1alpha1/frontendextensions`
- `/apis/<EXTENSION_API_GROUP>/<EXTENSION_API_VERSION>/frontendextensions`
- `/kapis/<EXTENSION_API_GROUP>/<EXTENSION_API_VERSION>/frontendextensions`

Default extension API group/version:

- `EXTENSION_API_GROUP=frontend-forge-api.kubesphere.io`
- `EXTENSION_API_VERSION=v1alpha1`

## Base URL

When operating through KubeSphere, use the KubeSphere API server origin and the `/kapis` path:

```bash
export KS_API=https://<kubesphere-host>
export FE_API="$KS_API/kapis/frontend-forge-api.kubesphere.io/v1alpha1/frontendextensions"
```

Use the user's existing browser/session token, kubeconfig proxy, or other environment-specific authentication method. Do not invent credentials.

If `/kapis` returns `401` or `403`, treat it as an authentication or KubeSphere authorization issue before falling back to raw Kubernetes operations. Show the user a command they can run with their own credentials:

```bash
curl -fS -u "user:password" "$FE_API/<name>"
curl -fS -u "user:password" "$FE_API/<name>/publish"
```

Use `curl -u "user:password"` only as an operator-facing example. Do not ask the user to send credentials back into the conversation.

For local cluster debugging, port-forward the FE API service:

```bash
kubectl -n extension-frontend-forge get svc -l app.kubernetes.io/component=extension-api
kubectl -n extension-frontend-forge port-forward svc/<extension-api-service> 18080:80
export FE_API=http://127.0.0.1:18080/kapis/frontend-forge-api.kubesphere.io/v1alpha1/frontendextensions
```

For a default Helm release named `frontend-forge`, the service is usually `frontend-forge-extension-api`.

If the KubeSphere APIService route is unavailable during local debugging, fall back to the direct resource route:

```bash
export FE_API=http://127.0.0.1:18080/apis/frontend-forge.kubesphere.io/v1alpha1/frontendextensions
```

## List And Get

```bash
curl -fS "$FE_API"
curl -fS -u "user:password" "$FE_API"
curl -fS "$FE_API?labelSelector=frontend-forge.kubesphere.io/package-state=ready,frontend-forge.kubesphere.io/publish-state=not-published"
curl -fS "$FE_API/<name>"
```

Use the publish/unpublish GET endpoints only to read operation status. They do not trigger publish or unpublish:

```bash
curl -fS "$FE_API/<name>/publish"
curl -fS "$FE_API/<name>/unpublish"
```

## Create

The API can create a cluster-scoped FE, but `kubectl apply -f` is usually better for GitOps or local samples.

```bash
curl -fS -X POST \
  -H 'Content-Type: application/json' \
  --data @frontendextension.json \
  "$FE_API"
```

## Download

```bash
curl -fS "$FE_API/<name>" | jq '.status.phase, .status.download, .status.artifact.digest'
curl -fL "$FE_API/<name>/download" -o <name>.tgz
```

Download returns `409` when the FE is not `Ready`, download is not ready, or the artifact does not match the current source hash.

## Publish

Check current publish status. This is read-only:

```bash
curl -fS "$FE_API/<name>/publish"
curl -fS -u "user:password" "$FE_API/<name>/publish"
```

Publish current artifact:

```bash
digest=$(kubectl get fe <name> -o jsonpath='{.status.artifact.digest}')
curl -fS -X POST \
  -H 'Content-Type: application/json' \
  --data '{"requestId":"manual-1"}' \
  "$FE_API/<name>/publish"
kubectl get fe <name> -o jsonpath='{.status.publish}{"\n"}'
```

Publish with digest protection. This is safer for manual operations because the API returns `409` if the artifact changed between inspection and publish:

```bash
digest=$(kubectl get fe <name> -o jsonpath='{.status.artifact.digest}')
curl -fS -X POST \
  -H 'Content-Type: application/json' \
  --data "{\"requestId\":\"manual-1\",\"expectedArtifactDigest\":\"${digest}\"}" \
  "$FE_API/<name>/publish"
curl -fS -u "user:password" -X POST \
  -H 'Content-Type: application/json' \
  --data "{\"requestId\":\"manual-1\",\"expectedArtifactDigest\":\"${digest}\"}" \
  "$FE_API/<name>/publish"
kubectl get fe <name> -o jsonpath='{.status.publish.phase}{" "}{.status.publish.active}{" "}{.status.publish.jobRef.namespace}{" "}{.status.publish.jobRef.name}{"\n"}'
```

The API patches publish annotations and returns `202 Accepted` when accepted. It returns `409` for artifact not ready, digest mismatch, missing target ref, or invalid target kind.

Follow the publisher Job if one is created:

```bash
job_ns=$(kubectl get fe <name> -o jsonpath='{.status.publish.jobRef.namespace}')
job_name=$(kubectl get fe <name> -o jsonpath='{.status.publish.jobRef.name}')
kubectl -n "$job_ns" get job "$job_name" -o yaml
kubectl -n "$job_ns" logs "job/$job_name"
```

## Unpublish

Check current unpublish status. This is read-only:

```bash
curl -fS "$FE_API/<name>/unpublish"
curl -fS -u "user:password" "$FE_API/<name>/unpublish"
```

Trigger unpublish:

```bash
curl -fS -X POST \
  -H 'Content-Type: application/json' \
  --data '{"requestId":"manual-unpublish-1"}' \
  "$FE_API/<name>/unpublish"
curl -fS -u "user:password" -X POST \
  -H 'Content-Type: application/json' \
  --data '{"requestId":"manual-unpublish-1"}' \
  "$FE_API/<name>/unpublish"
kubectl get fe <name> -o jsonpath='{.status.unpublish.phase}{" "}{.status.unpublish.jobRef.namespace}{" "}{.status.unpublish.jobRef.name}{"\n"}'
```

The API patches unpublish annotations and returns `202 Accepted` when accepted.

Follow the unpublish Job if one is created:

```bash
job_ns=$(kubectl get fe <name> -o jsonpath='{.status.unpublish.jobRef.namespace}')
job_name=$(kubectl get fe <name> -o jsonpath='{.status.unpublish.jobRef.name}')
kubectl -n "$job_ns" get job "$job_name" -o yaml
kubectl -n "$job_ns" logs "job/$job_name"
```

## Delete

Direct delete without unpublish:

```bash
curl -fS -X POST \
  -H 'Content-Type: application/json' \
  --data '{"unpublish":false}' \
  "$FE_API/<name>/delete"
```

Unpublish first if currently published:

```bash
kubectl get fe <name> -o jsonpath='{.status.publish.phase}{" "}{.status.publish.active}{" "}{.status.publish.artifactDigest}{"\n"}'
curl -fS -X POST \
  -H 'Content-Type: application/json' \
  --data '{"unpublish":true}' \
  "$FE_API/<name>/delete"
curl -fS -u "user:password" -X POST \
  -H 'Content-Type: application/json' \
  --data '{"unpublish":true}' \
  "$FE_API/<name>/delete"
```

Expected behavior:

- `200 OK` when the FE is directly deleted
- `202 Accepted` when an active published FE needs unpublish first

Delete with unpublish example:

```bash
export FE_API="$KS_API/kapis/frontend-forge-api.kubesphere.io/v1alpha1/frontendextensions"
curl -fS "$FE_API/<name>/publish"
curl -fS -X POST \
  -H 'Content-Type: application/json' \
  --data '{"unpublish":true}' \
  "$FE_API/<name>/delete"
curl -fS "$FE_API/<name>/unpublish"
```

If KubeSphere requires basic auth for this route, the same flow can be shown to the user as:

```bash
curl -fS -u "user:password" "$FE_API/<name>/publish"
curl -fS -u "user:password" -X POST \
  -H 'Content-Type: application/json' \
  --data '{"unpublish":true}' \
  "$FE_API/<name>/delete"
curl -fS -u "user:password" "$FE_API/<name>/unpublish"
```

After a `202 Accepted` response, follow the unpublish Job from FE status until it succeeds and the FE is deleted.

## API Response And Status Mapping

- `GET /<name>/publish` is read-only and returns `status.publish` or default `NotRequested`.
- `GET /<name>/unpublish` is read-only and returns `status.unpublish` or default `NotRequested`.
- `POST /<name>/publish` returns `202 Accepted` and a `PublishStatus` with `phase=Pending` unless it is idempotently returning an existing matching status.
- `POST /<name>/unpublish` returns `202 Accepted` and an `UnpublishStatus` with `phase=Pending` unless it is idempotently returning an existing matching status.
- `POST /<name>/delete` returns `200 OK` with `deleted=true` for direct deletion, or `202 Accepted` with `deleted=false` when unpublish must run first.

## Raw Annotation Patch Fallback Is Last Resort

Publish and unpublish annotations are implementation details of the API/controller handshake. Inspect them when debugging, but do not use raw annotation patches as the normal operation path.

Use a raw annotation patch only when all of these are true:

- the FE API is unavailable
- the user explicitly asks for an annotation-based recovery or accepts the risk
- current `.metadata.generation`, `.status.observedSourceHash`, `.status.artifact.digest`, and target ref have been re-read immediately before patching
- the patch is scoped to one named FE

Raw patch pitfalls:

- stale generation/source-hash annotations can be ignored or marked failed
- missing artifact digest can fail publish reconciliation
- target ref must resolve to a `ConfigMap` or `Secret`
- delete-after-unpublish requires `frontend-forge.kubesphere.io/delete-after-unpublish-request-id`
