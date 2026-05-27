# FrontendExtension Lifecycle Reference

Primary sources:

- `config/samples/frontendextension-inspecttask.yaml`
- `spec/frontend-extension-design.md`
- `spec/crds.md`
- live `kubectl get fe <name> -o yaml`

## Create

Apply an FE manifest:

```bash
kubectl apply -f config/samples/frontendextension-inspecttask.yaml
kubectl get fe inspecttask -o yaml
```

Expected progression:

- `status.phase` starts as `Pending` or `Packaging`
- a package Job is created in the controller work namespace
- a matching artifact ConfigMap is written
- `status.phase` becomes `Ready`
- `status.download.ready` becomes `true`
- labels reflect package state:
  - `frontend-forge.kubesphere.io/package-state=ready`
  - `frontend-forge.kubesphere.io/publish-state=not-published`

## Update

Edit the manifest and re-apply:

```bash
kubectl apply -f <file.yaml>
kubectl get fe <name> -o yaml
```

After a source or package identity change, check:

- `.metadata.generation`
- `.status.observedGeneration`
- `.status.observedSourceHash`
- `.status.artifact.artifactKey`
- `.status.packageJob.name`
- `.status.conditions`

The source hash includes package/source identity and excludes publish policy. Changing only `spec.publishPolicy` should not require a new package artifact.

## Force Rebuild

Use the rebuild-token annotation when the source is unchanged but a fresh package is needed:

```bash
kubectl get fe <name> -o jsonpath='{.status.observedSourceHash}{" "}{.status.artifact.digest}{" "}{.status.artifact.artifactKey}{" "}{.status.publish.phase}{" "}{.status.publish.active}{"\n"}'
kubectl annotate fe <name> frontend-forge.kubesphere.io/rebuild-token="$(date +%s)" --overwrite
kubectl get fe <name> -o yaml
```

Check:

- `.status.observedRebuildToken`
- `.status.artifact.artifactKey`
- `.status.packageJob.name`
- package Job attempt and logs

The rebuild token changes artifact cache identity. It does not change package content by itself unless the build service or source inputs produce different bytes.

Rebuild safety:

- Confirm the user wants a new package artifact, not only publish/unpublish status refresh.
- Record the previous artifact digest and publish status because a rebuild can make the active publish stale.
- After the rebuild, wait for `status.phase=Ready` before download or publish.

## Download

Prefer the FE HTTP API because it verifies readiness, source hash, artifact storage, and digest.

Port-forward the service if needed:

```bash
kubectl -n extension-frontend-forge port-forward svc/frontend-forge-extension-api 18080:80
```

Then download:

```bash
KS_API=https://<kubesphere-host>
FE_API="$KS_API/kapis/frontend-forge-api.kubesphere.io/v1alpha1/frontendextensions"
curl -fL "$FE_API/<name>/download" -o <name>.tgz
curl -fL -u "user:password" "$FE_API/<name>/download" -o <name>.tgz
```

If the service name differs, discover it:

```bash
kubectl -n extension-frontend-forge get svc -l app.kubernetes.io/component=extension-api
```

## Delete

Direct Kubernetes delete removes the FE CR and does not automatically unpublish:

```bash
kubectl delete fe <name>
```

Use the FE API delete endpoint when a published extension should be unpublished before deletion:

```bash
kubectl get fe <name> -o jsonpath='{.status.publish.phase}{" "}{.status.publish.active}{" "}{.status.publish.artifactDigest}{"\n"}'
KS_API=https://<kubesphere-host>
FE_API="$KS_API/kapis/frontend-forge-api.kubesphere.io/v1alpha1/frontendextensions"
curl -fS -X POST \
  -H 'Content-Type: application/json' \
  --data '{"unpublish":true}' \
  "$FE_API/<name>/delete"
curl -fS -u "user:password" -X POST \
  -H 'Content-Type: application/json' \
  --data '{"unpublish":true}' \
  "$FE_API/<name>/delete"
curl -fS "$FE_API/<name>/unpublish"
```

If the extension is currently published, expect `202 Accepted` and an unpublish Job. If it is not currently published, expect direct deletion.

Deletion safety:

- Use API delete with `{"unpublish":true}` for published extensions.
- Use direct `kubectl delete fe <name>` only when the user wants to remove the CR without running unpublish.
- Do not delete package or publisher Jobs while they are active unless the user is intentionally aborting the operation.

## Expected Package Phases

- `Pending`: default before the controller writes status
- `Packaging`: package Job created or running
- `Ready`: matching artifact ConfigMap exists and metadata matches current source hash and artifact key
- `Failed`: invalid source, package attempts exceeded, or artifact missing/mismatched after package attempts
