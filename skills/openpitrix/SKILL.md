---
name: openpitrix
description: KubeSphere OpenPitrix application management Skill. Use when users ask about KubeSphere App Store, OpenPitrix, Helm/YAML application templates, application repositories, app versions, app releases, categories, review states, repository sync, or troubleshooting application installation and upgrade issues.
---

# OpenPitrix Application Management

## Overview

OpenPitrix is KubeSphere's application management capability for app repositories, app templates, versions, reviews, and app releases. In KubeSphere 4.x the runtime API group is `application.kubesphere.io/v2`; older OpenPitrix extension code used `openpitrix.io/v2` CRUD APIs and `openpitrix.io/v2alpha1` read/list wrappers over `application.kubesphere.io/v1alpha1` resources.

Use the v2 objects and APIs first:

| Concept | KubeSphere 4.x object | Older OpenPitrix object |
|---|---|---|
| Repository | `Repo` | `HelmRepo` |
| App template | `Application` | `HelmApplication` |
| App template version | `ApplicationVersion` | `HelmApplicationVersion` |
| App release / installed app | `ApplicationRelease` | `HelmRelease` |
| Category | `Category` | `HelmCategory` |

Core namespace and labels:

| Item | Value |
|---|---|
| Application data namespace | `extension-openpitrix` |
| Repository label | `application.kubesphere.io/repo-name` |
| App label | `application.kubesphere.io/app-id` |
| App version label | `application.kubesphere.io/appversion-id` |
| App type label | `application.kubesphere.io/app-type` |
| Cluster label | `kubesphere.io/cluster` |
| Namespace label | `kubesphere.io/namespace` |
| Workspace label | `kubesphere.io/workspace` |

## Architecture

```
Helm repo index or uploaded package
        |
        v
Repo sync / upload API
        |
        v
Application -> ApplicationVersion -> ApplicationRelease
                                      |
                                      v
                         Helm executor Job or YAML installer
                                      |
                                      v
                          Workloads in target cluster/namespace
```

Important controllers:

| Controller | Watches | Purpose |
|---|---|---|
| `helmrepo-controller` | `Repo` | Loads Helm repository indexes, creates/deletes `Application` and `ApplicationVersion`, updates repository sync state. |
| `appversion-controller` | `ApplicationVersion` | Cleans stored chart/YAML data after version deletion when it is no longer used. |
| `apprelease-helminstaller` | `ApplicationRelease` | Creates, upgrades, verifies, and uninstalls app releases through Helm executor Jobs or YAML installer logic. |
| `appcategory-controller` | `Category` | Maintains category counts and prevents deleting categories that still own apps. |

## Navigation and Feature Coverage

When the user says "应用商店", first identify whether they mean the enterprise-space app management pages or the global component-dock App Store management extension. They share the same OpenPitrix/KSE v2 resources, but the intent and scope differ.

Enterprise-space application management under a workspace:

| Console area | Typical route | User intent | Main resource/API |
|---|---|---|---|
| 应用管理 / 应用 | `/workspaces/{workspace}/deploy` | List, create, edit, upgrade, or delete installed apps in projects. | `ApplicationRelease`; `/workspaces/{workspace}/applications`, `/namespaces/{namespace}/applications` |
| 应用管理 / 自制应用 | workspace custom app area | Work with user-created/custom applications before or outside App Store publication. | Usually app template/upload flows; verify against `Application` and `ApplicationVersion` before assuming release APIs. |
| 应用管理 / 应用模板 | `/workspaces/{workspace}/app-templates` | Create/upload Helm or YAML app templates, edit template metadata, submit versions for review, manage versions. | `Application`, `ApplicationVersion`; `/workspaces/{workspace}/apps`, `/workspaces/{workspace}/apps/{app}/versions` |
| 应用管理 / 应用仓库 | `/workspaces/{workspace}/app-repos` | Add, sync, inspect, or delete Helm repos available to the workspace. | `Repo`; `/workspaces/{workspace}/repos` |

Component-dock App Store management:

| Console area | Typical route | User intent | Main resource/API |
|---|---|---|---|
| 组件坞 / 应用商店管理 / 应用 | `/apps-manage/store` | Platform-level App Store template list, publish/unpublish, edit metadata, delete, open detail pages. | `Application`, `ApplicationVersion`; `/workspaces/{workspace}/apps`, `/apps/{app}/action` |
| 组件坞 / 应用商店管理 / 应用分类 | `/apps-manage/categories` | Create, edit, delete categories and assign applications to categories. | `Category`; `/categories`, `/workspaces/{workspace}/apps/{app}` |
| 组件坞 / 应用商店管理 / 应用审核 | `/apps-manage/reviews` | Review uploaded app versions submitted from enterprise spaces. | `ApplicationVersion`; `/reviews`, `/workspaces/{workspace}/apps/{app}/versions/{version}/action` |
| 组件坞 / 应用商店管理 / 应用仓库 | `/apps-manage/repo` | Manage global/platform view of App Store repositories. | `Repo`; `/workspaces/{workspace}/repos` |
| 组件坞 / 应用商店管理 / 部署管理 | `/apps-manage/deploy` | Manage installed app releases across workspace/cluster/namespace scope. | `ApplicationRelease`; `/workspaces/{workspace}/applications`, `/namespaces/{namespace}/applications` |

Resource selection rules:

- Use `Application` and `ApplicationVersion` for app templates, App Store listings, uploaded packages, version review, screenshots, metadata, and categories.
- Use `Repo` for app repositories and repository sync.
- Use `Category` for App Store category management.
- Use `ApplicationRelease` only for installed/deployed apps, including workspace "应用" pages and platform "部署管理" pages.
- Do not troubleshoot an App Store template with release Job/Pod commands unless the user is installing or upgrading an `ApplicationRelease`.
- Uploaded/self-made app templates use the repository label value `application.kubesphere.io/repo-name=upload`. Do not write `uploaded`.
- KSE v2 app and version action examples use `{"state":"..."}` with optional `message`, not legacy `{"action":"..."}`. Legacy `openpitrix.io/v2` action APIs use `action`.
- For `kubectl get/describe` of OpenPitrix application CRDs, do not add `-n extension-openpitrix` by default. These resources are queried by resource kind and labels; use labels such as `kubesphere.io/workspace`, `application.kubesphere.io/repo-name`, `application.kubesphere.io/app-id`, or `application.kubesphere.io/app-release-name`. Use `extension-openpitrix` only when inspecting extension component Pods or storage fallback objects.

## Tool Selection

Choose the tool by task. Prefer the first matching option:

| Tool | Use for | Authentication |
|---|---|---|
| `kubectl` | Inspect CRDs/resources, events, controller state, executor Jobs/Pods/logs, and cluster-side troubleshooting. | Uses the current kubeconfig. |
| `ks_api.py` | KubeSphere JSON KAPIs under `/kapis/...`; recommended for create/update/list/action calls that send JSON. | Run login once; token is cached in `~/.kubesphere_token`. |
| `curl` | Multipart uploads, package/file downloads, custom headers, reproducing exact HTTP requests, or when the user explicitly asks for curl. | Requires `Authorization: Bearer $TOKEN`; get the token with `ks_api.py --login` or set it manually. |

Set up `ks_api.py` first when using KubeSphere KAPIs:

```bash
cd skills/kubesphere-core/scripts
export KUBESPHERE_HOST="http://<kubesphere-host>"
python ks_api.py --login --username admin --password <password>
```

For curl, reuse the token cached by `ks_api.py`:

```bash
export KUBESPHERE_HOST="http://<kubesphere-host>"
export TOKEN=$(python -c 'import json, os; print(json.load(open(os.path.expanduser("~/.kubesphere_token")))["token"])')
```

If `ks_api.py` is unavailable, obtain an OAuth token directly:

```bash
export KUBESPHERE_HOST="http://<kubesphere-host>"
export TOKEN=$(curl -sS -X POST "$KUBESPHERE_HOST/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "username=<username>" \
  -d "password=<password>" \
  -d "client_id=kubesphere" \
  -d "client_secret=kubesphere" | jq -r '.access_token')
```

Then pass `-H "Authorization: Bearer $TOKEN"` on every curl request to `/kapis/...`.

## Quick Inspection

Start with cluster resources before calling KAPIs:

```bash
kubectl get repos.application.kubesphere.io
kubectl get applications.application.kubesphere.io
kubectl get applicationversions.application.kubesphere.io
kubectl get applicationreleases.application.kubesphere.io
kubectl get categories.application.kubesphere.io
```

For workspace-scoped views, filter by workspace label:

```bash
kubectl get repos.application.kubesphere.io \
  -l kubesphere.io/workspace=<workspace>

kubectl get applications.application.kubesphere.io \
  -l kubesphere.io/workspace=<workspace>
```

For an installed app:

```bash
kubectl get applicationreleases.application.kubesphere.io \
  -l kubesphere.io/cluster=<cluster>,kubesphere.io/namespace=<namespace>

kubectl describe applicationrelease.application.kubesphere.io <release-name>
```

## KAPI Routes

Use `/kapis/application.kubesphere.io/v2` in KubeSphere 4.x.

Repository routes:

| Operation | Route |
|---|---|
| List repositories | `GET /workspaces/{workspace}/repos` |
| Create repository | `POST /workspaces/{workspace}/repos` |
| Update repository | `PATCH /workspaces/{workspace}/repos/{repo}` |
| Delete repository | `DELETE /workspaces/{workspace}/repos/{repo}` |
| Manual sync | `POST /workspaces/{workspace}/repos/{repo}/action` |
| Repository events | `GET /workspaces/{workspace}/repos/{repo}/events` |

App template routes:

| Operation | Route |
|---|---|
| List apps | `GET /workspaces/{workspace}/apps` |
| Create uploaded app | `POST /workspaces/{workspace}/apps` |
| Describe app | `GET /workspaces/{workspace}/apps/{app}` |
| Create/update app metadata | `POST /workspaces/{workspace}/apps/{app}` |
| Patch metadata | `PATCH /workspaces/{workspace}/apps/{app}` |
| Delete app | `DELETE /workspaces/{workspace}/apps/{app}` |
| Review/action app | `POST /apps/{app}/action` |

Version and release routes:

| Operation | Route |
|---|---|
| List versions | `GET /workspaces/{workspace}/apps/{app}/versions` |
| Create version | `POST /workspaces/{workspace}/apps/{app}/versions` |
| Describe version | `GET /workspaces/{workspace}/apps/{app}/versions/{version}` |
| Download package | `GET /workspaces/{workspace}/apps/{app}/versions/{version}/package` |
| List chart/YAML files | `GET /workspaces/{workspace}/apps/{app}/versions/{version}/files` |
| Review/action version | `POST /workspaces/{workspace}/apps/{app}/versions/{version}/action` |
| List releases by workspace | `GET /workspaces/{workspace}/applications` |
| List releases by namespace | `GET /namespaces/{namespace}/applications` |
| Create release | `POST /namespaces/{namespace}/applications` |
| Describe release | `GET /namespaces/{namespace}/applications/{application}` |
| Delete release | `DELETE /namespaces/{namespace}/applications/{application}` |

App Store management routes:

| Operation | Route |
|---|---|
| List categories | `GET /categories` |
| Create category | `POST /categories` |
| Update category | `POST /categories/{category}` |
| Describe category | `GET /categories/{category}` |
| Delete category | `DELETE /categories/{category}` |
| List app reviews | `GET /reviews` |
| Upload attachment | `POST /workspaces/{workspace}/attachments` |
| Describe attachment | `GET /workspaces/{workspace}/attachments/{attachment}` |
| Delete attachments | `DELETE /workspaces/{workspace}/attachments/{attachment}` |

Older OpenPitrix extension routes use `/kapis/openpitrix.io/v2alpha1`, primarily as read/list wrappers around `HelmRepo`, `HelmApplication`, `HelmApplicationVersion`, `HelmRelease`, and `HelmCategory`.

Assume `KUBESPHERE_HOST`, `ks_api.py` login, and `TOKEN` have already been set up from [Tool Selection](#tool-selection) before using the examples below.

### KSE Application API examples

Use these examples for KubeSphere 4.x `application.kubesphere.io/v2`.

```bash
# Create or update repository.
python ks_api.py POST /kapis/application.kubesphere.io/v2/workspaces/<workspace>/repos '{
  "metadata": {
    "name": "<repo-name>",
    "labels": {
      "kubesphere.io/workspace": "<workspace>"
    },
    "annotations": {
      "kubesphere.io/display-name": "<display-name>"
    }
  },
  "spec": {
    "url": "https://example.com/charts",
    "description": "<description>",
    "syncPeriod": 0
  }
}'

# Manually trigger repository sync.
python ks_api.py POST \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/repos/<repo-name>/action

# List apps in a workspace.
python ks_api.py GET \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps

# Describe an app and list versions.
python ks_api.py GET \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps/<app>

python ks_api.py GET \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps/<app>/versions

# Review or publish an app version.
python ks_api.py POST \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps/<app>/versions/<version>/action \
  '{"state":"active","message":"publish"}'

# App-level publish/suspend/recover actions also use state, not legacy action.
python ks_api.py POST \
  /kapis/application.kubesphere.io/v2/apps/<app>/action \
  '{"state":"suspended","message":"suspend from App Store"}'

# Create or update an application release.
# Important:
# - Use /namespaces/<namespace>/applications, not /workspaces/<workspace>/namespaces/<namespace>/applications.
# - Required release references are spec.appID, spec.appVersionID, spec.appType, and labels.
# - Do not invent spec.name or spec.namespace; release name and namespace are metadata/path concerns.
# - spec.values is a JSON []byte field: use "" for empty values, or base64-encoded YAML bytes for non-empty values. Do not use {}.
python ks_api.py POST \
  /kapis/application.kubesphere.io/v2/namespaces/<namespace>/applications '{
    "metadata": {
      "name": "<release-name>",
      "labels": {
        "application.kubesphere.io/app-id": "<app-id>",
        "application.kubesphere.io/appversion-id": "<app-version-id>",
        "application.kubesphere.io/app-type": "helm",
        "kubesphere.io/cluster": "<cluster>",
        "kubesphere.io/namespace": "<namespace>",
        "kubesphere.io/workspace": "<workspace>"
      },
      "annotations": {
        "kubesphere.io/creator": "<username>"
      }
    },
    "spec": {
      "appID": "<app-id>",
      "appVersionID": "<app-version-id>",
      "appType": "helm",
      "values": ""
    }
  }'

# List or describe releases in a namespace.
python ks_api.py GET \
  /kapis/application.kubesphere.io/v2/namespaces/<namespace>/applications

# Or list releases by workspace.
python ks_api.py GET \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/applications

python ks_api.py GET \
  /kapis/application.kubesphere.io/v2/namespaces/<namespace>/applications/<release-name>

# Categories and reviews.
python ks_api.py GET /kapis/application.kubesphere.io/v2/categories
python ks_api.py GET /kapis/application.kubesphere.io/v2/reviews
```

Use curl only when the API needs multipart upload, streaming download, or custom headers that `ks_api.py` does not support.

Curl equivalents for JSON KAPIs:

```bash
export KUBESPHERE_HOST="http://<kubesphere-host>"
export TOKEN="<kubesphere-access-token>"

# Create or update repository.
curl -sS -X POST \
  "$KUBESPHERE_HOST/kapis/application.kubesphere.io/v2/workspaces/<workspace>/repos" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "name": "<repo-name>",
      "labels": {
        "kubesphere.io/workspace": "<workspace>"
      },
      "annotations": {
        "kubesphere.io/display-name": "<display-name>"
      }
    },
    "spec": {
      "url": "https://example.com/charts",
      "description": "<description>",
      "syncPeriod": 0
    }
  }'

# Manually trigger repository sync.
curl -sS -X POST \
  "$KUBESPHERE_HOST/kapis/application.kubesphere.io/v2/workspaces/<workspace>/repos/<repo-name>/action" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'

# For KSE v2 repository sync, use an empty JSON body or omit the body.
# Do not send legacy {"action":"sync"} or {"action":"index"} unless using openpitrix.io/v2.

# List apps in a workspace.
curl -sS \
  "$KUBESPHERE_HOST/kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps" \
  -H "Authorization: Bearer $TOKEN"

# Publish or review an app version.
curl -sS -X POST \
  "$KUBESPHERE_HOST/kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps/<app>/versions/<version>/action" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"state":"active","message":"publish"}'

# Create or update an application release.
# Keep the body shape aligned with ApplicationReleaseSpec. Do not add spec.name/spec.namespace.
# Use values: "" for empty values; do not use values: {}.
curl -sS -X POST \
  "$KUBESPHERE_HOST/kapis/application.kubesphere.io/v2/namespaces/<namespace>/applications" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "name": "<release-name>",
      "labels": {
        "application.kubesphere.io/app-id": "<app-id>",
        "application.kubesphere.io/appversion-id": "<app-version-id>",
        "application.kubesphere.io/app-type": "helm",
        "kubesphere.io/cluster": "<cluster>",
        "kubesphere.io/namespace": "<namespace>",
        "kubesphere.io/workspace": "<workspace>"
      }
    },
    "spec": {
      "appID": "<app-id>",
      "appVersionID": "<app-version-id>",
      "appType": "helm",
      "values": ""
    }
  }'
```

### Legacy OpenPitrix API examples

Use these only when the installed OpenPitrix extension exposes the old `openpitrix.io` KAPIs. Prefer `application.kubesphere.io/v2` for KubeSphere 4.x. The legacy request/response fields use OpenPitrix-style snake_case names such as `repo_id`, `version_id`, `sync_period`, and `app_default_status`.

Read/list wrappers from `/kapis/openpitrix.io/v2alpha1`:

```bash
# Repositories.
python ks_api.py GET \
  /kapis/openpitrix.io/v2alpha1/workspaces/<workspace>/repos

python ks_api.py GET \
  /kapis/openpitrix.io/v2alpha1/workspaces/<workspace>/repos/<repo-id>

# App templates and versions.
python ks_api.py GET \
  /kapis/openpitrix.io/v2alpha1/workspaces/<workspace>/apps

python ks_api.py GET \
  /kapis/openpitrix.io/v2alpha1/workspaces/<workspace>/apps/<app-id>

python ks_api.py GET \
  /kapis/openpitrix.io/v2alpha1/workspaces/<workspace>/apps/<app-id>/versions

python ks_api.py GET \
  /kapis/openpitrix.io/v2alpha1/workspaces/<workspace>/apps/<app-id>/versions/<version-id>

# Installed applications.
python ks_api.py GET \
  /kapis/openpitrix.io/v2alpha1/workspaces/<workspace>/clusters/<cluster>/namespaces/<namespace>/applications

python ks_api.py GET \
  /kapis/openpitrix.io/v2alpha1/workspaces/<workspace>/clusters/<cluster>/namespaces/<namespace>/applications/<application-id>

# Categories.
python ks_api.py GET /kapis/openpitrix.io/v2alpha1/categories
python ks_api.py GET /kapis/openpitrix.io/v2alpha1/categories/<category-id>
```

Older CRUD-style APIs from `/kapis/openpitrix.io/v2`:

```bash
# Create repository. Use validate=true to validate without persisting.
python ks_api.py POST \
  /kapis/openpitrix.io/v2/workspaces/<workspace>/repos?validate=true '{
    "name": "<repo-name>",
    "url": "https://example.com/charts",
    "type": "helm",
    "visibility": "public",
    "providers": ["kubernetes"],
    "sync_period": "0s",
    "app_default_status": "active",
    "credential": ""
  }'

# Trigger repository indexing/sync.
python ks_api.py POST \
  /kapis/openpitrix.io/v2/workspaces/<workspace>/repos/<repo-id>/action \
  '{"action":"index","workspace":"<workspace>"}'

# Create an app template from a base64-encoded chart package.
python ks_api.py POST \
  /kapis/openpitrix.io/v2/workspaces/<workspace>/apps '{
    "name": "<app-name>",
    "version_name": "0.1.0",
    "version_type": "helm",
    "version_package": "<base64-chart-tgz>"
  }'

# Create another version for an existing app.
python ks_api.py POST \
  /kapis/openpitrix.io/v2/workspaces/<workspace>/apps/<app-id>/versions '{
    "app_id": "<app-id>",
    "name": "0.2.0",
    "type": "helm",
    "package": "<base64-chart-tgz>"
  }'

# Submit, pass, reject, suspend, recover, or activate an app version.
python ks_api.py POST \
  /kapis/openpitrix.io/v2/workspaces/<workspace>/apps/<app-id>/versions/<version-id>/action \
  '{"action":"submit","message":"submit for review"}'

# Deploy an app release.
python ks_api.py POST \
  /kapis/openpitrix.io/v2/workspaces/<workspace>/clusters/<cluster>/namespaces/<namespace>/applications '{
    "name": "<release-name>",
    "app_id": "<app-id>",
    "version_id": "<version-id>",
    "runtime_id": "<cluster>",
    "conf": "{}",
    "advanced_param": []
  }'
```

## App Store and Workspace App Management Workflow

KubeSphere's OpenPitrix extension serves both enterprise-space application management and component-dock App Store management. Map the console area to the KSE v2 APIs before choosing commands:

| UI area | Main purpose | Primary APIs |
|---|---|---|
| `/workspaces/{workspace}/deploy` | Enterprise-space installed apps. | `/workspaces/{workspace}/applications`, `/namespaces/{namespace}/applications` |
| `/workspaces/{workspace}/app-templates` | Enterprise-space app templates created from Helm/YAML packages. | `/workspaces/{workspace}/apps`, `/workspaces/{workspace}/apps/{app}/versions`, `/workspaces/{workspace}/attachments` |
| `/workspaces/{workspace}/app-repos` | Enterprise-space app repositories. | `/workspaces/{workspace}/repos` |
| `/apps-manage/store` | Manage application templates in the App Store: list, create/upload, edit metadata, delete, open detail pages. | `/workspaces/{workspace}/apps`, `/workspaces/{workspace}/apps/{app}`, `/workspaces/{workspace}/attachments` |
| `/apps-manage/store/{app}` | Inspect template details, versions, audit records, and deployed instances. | `/workspaces/{workspace}/apps/{app}`, `/workspaces/{workspace}/apps/{app}/versions`, `/workspaces/{workspace}/applications` |
| `/apps-manage/categories` | Manage categories and assign apps to categories. | `/categories`, `/categories/{category}`, `/workspaces/{workspace}/apps/{app}` |
| `/apps-manage/reviews` | Review uploaded app versions. | `/reviews`, `/workspaces/{workspace}/apps/{app}/versions/{version}/action` |
| `/apps-manage/repo` | Manage Helm repositories that feed App Store templates. | `/workspaces/{workspace}/repos` |
| `/apps-manage/deploy` | Manage installed app releases. | `/namespaces/{namespace}/applications`, `/workspaces/{workspace}/applications` |

List and filter behavior:

- App template and App Store pages list `Application` templates, not installed `ApplicationRelease` objects.
- Workspace "应用" and App Store "部署管理" pages list installed `ApplicationRelease` objects.
- The UI filters list queries through KubeSphere list query parameters such as `conditions`, `status`, `order`, `limit`, and workspace query scope.
- Public store display commonly focuses on `active|suspended` apps; management views include draft, passed, active, and suspended states.
- Uploaded apps are identified with `application.kubesphere.io/repo-name=upload`.

Inspect app templates:

```bash
python ks_api.py GET \
  "/kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps?conditions=status=draft|active|suspended|passed&sortBy=create_time"

kubectl get applications.application.kubesphere.io \
  -l kubesphere.io/workspace=<workspace>
```

Inspect installed apps/releases:

```bash
python ks_api.py GET \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/applications

python ks_api.py GET \
  /kapis/application.kubesphere.io/v2/namespaces/<namespace>/applications

kubectl get applicationreleases.application.kubesphere.io \
  -l kubesphere.io/workspace=<workspace>
```

Patch app template metadata. This is how the enterprise-space template page and App Store management page edit alias, description, icon, category, screenshots/attachments, abstraction, and home URL:

```bash
python ks_api.py PATCH \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps/<app> '{
    "aliasName": "<display-name>",
    "description": "<description>",
    "categoryName": "<category>",
    "icon": "<base64-icon-or-url>",
    "attachments": ["<attachment-id>"],
    "abstraction": "<short-summary>",
    "appHome": "https://example.com"
  }'
```

The patch handler writes these fields to:

| Request field | Stored as |
|---|---|
| `categoryName` | `metadata.labels["application.kubesphere.io/app-category-name"]` |
| `aliasName` | `metadata.annotations["kubesphere.io/display-name"]` |
| `description` | `metadata.annotations["kubesphere.io/description"]` |
| `icon` | `spec.icon` |
| `attachments` | `spec.attachments` |
| `abstraction` | `spec.abstraction` |
| `appHome` | `spec.appHome` |

Manage attachments for App Store screenshots and other assets. This API is multipart, so prefer curl with the bearer token:

```bash
curl -sS -X POST \
  "$KUBESPHERE_HOST/kapis/application.kubesphere.io/v2/workspaces/<workspace>/attachments" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@./screenshot.png"

curl -sS \
  "$KUBESPHERE_HOST/kapis/application.kubesphere.io/v2/workspaces/<workspace>/attachments/<attachment-id>" \
  -H "Authorization: Bearer $TOKEN"

curl -sS -X DELETE \
  "$KUBESPHERE_HOST/kapis/application.kubesphere.io/v2/workspaces/<workspace>/attachments/<attachment-id>" \
  -H "Authorization: Bearer $TOKEN"
```

Manage categories:

```bash
python ks_api.py GET /kapis/application.kubesphere.io/v2/categories

python ks_api.py POST /kapis/application.kubesphere.io/v2/categories '{
  "metadata": {
    "name": "<category>",
    "annotations": {
      "kubesphere.io/display-name": "<display-name>",
      "kubesphere.io/description": "<description>"
    }
  },
  "spec": {
    "icon": "database"
  }
}'

python ks_api.py POST /kapis/application.kubesphere.io/v2/categories/<category> '{
  "metadata": {
    "name": "<category>",
    "annotations": {
      "kubesphere.io/display-name": "<display-name>",
      "kubesphere.io/description": "<description>"
    }
  },
  "spec": {
    "icon": "database"
  }
}'

python ks_api.py DELETE /kapis/application.kubesphere.io/v2/categories/<category>
```

Do not delete `kubesphere-app-uncategorized`, and do not delete a category whose `status.total` is greater than zero. To move apps between categories, patch each app's `categoryName` through `/workspaces/{workspace}/apps/{app}`.

Review uploaded app versions:

```bash
python ks_api.py GET \
  "/kapis/application.kubesphere.io/v2/reviews?conditions=status=submitted"

python ks_api.py GET \
  "/kapis/application.kubesphere.io/v2/reviews?conditions=status=active|rejected|passed|submitted|suspended"

python ks_api.py POST \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps/<app>/versions/<version>/action \
  '{"state":"passed","message":"approve"}'

python ks_api.py POST \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps/<app>/versions/<version>/action \
  '{"state":"rejected","message":"reject reason"}'
```

`/reviews` lists uploaded app versions only; the handler selects versions whose repo label is `application.kubesphere.io/repo-name=upload`. If a review item is missing, first verify that the `ApplicationVersion` belongs to the upload repo and is in a review state such as `submitted`.

Important KSE v2 App Store API shapes:

```bash
# Correct uploaded/self-made template label.
kubectl get applicationversions.application.kubesphere.io \
  -l application.kubesphere.io/repo-name=upload

# Correct category create/update body is a Category object.
python ks_api.py POST /kapis/application.kubesphere.io/v2/categories '{
  "metadata": {
    "name": "<category>",
    "annotations": {
      "kubesphere.io/display-name": "<display-name>",
      "kubesphere.io/description": "<description>"
    }
  },
  "spec": {
    "icon": "database"
  }
}'

# Correct repository create/update body is a Repo object shape.
python ks_api.py POST /kapis/application.kubesphere.io/v2/workspaces/<workspace>/repos '{
  "metadata": {
    "name": "<repo-name>",
    "labels": {
      "kubesphere.io/workspace": "<workspace>"
    },
    "annotations": {
      "kubesphere.io/display-name": "<display-name>"
    }
  },
  "spec": {
    "url": "https://example.com/charts",
    "description": "<description>",
    "syncPeriod": 0
  }
}'

# Correct KSE v2 review/action body uses state.
python ks_api.py POST \
  /kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps/<app>/versions/<version>/action \
  '{"state":"passed","message":"approve"}'
```

Avoid these common mistakes for `application.kubesphere.io/v2`:

- Do not use `application.kubesphere.io/repo-name=uploaded`; the built-in upload repo key is `upload`.
- Do not send top-level category bodies such as `{"name":"...","displayName":"..."}`; send a `Category` object with `metadata` and `spec`.
- Do not send top-level repo bodies such as `{"name":"...","url":"..."}`; send a `Repo` object with `metadata` and `spec`.
- Do not send `{"action":"approve"}`, `{"action":"reject"}`, or `{"action":"sync"}` to KSE v2 action routes; use `state` for app/version actions and an empty body for repository manual sync.

## Repository Workflow

Create or update a repository with a valid Helm repository URL. The API validates the URL by loading the repository index before persisting it. User info embedded in the URL is copied into `spec.credential`.

```yaml
apiVersion: application.kubesphere.io/v2
kind: Repo
metadata:
  name: <repo-name>
  labels:
    kubesphere.io/workspace: <workspace>
  annotations:
    kubesphere.io/display-name: <display-name>
spec:
  url: https://example.com/charts
  description: <description>
  syncPeriod: 0
```

Sync behavior:

- `spec.syncPeriod: 0` means no periodic sync.
- Manual sync sets `status.state` to `manualTrigger`.
- Successful sync sets `status.state` to `successful`.
- Repo sync creates app IDs as `<repo-name>-<short-hash-of-chart-name>`.
- Repo versions become active automatically because they came from a trusted repository.

Troubleshoot repository sync:

```bash
kubectl describe repo.application.kubesphere.io <repo-name>
kubectl get events --field-selector involvedObject.name=<repo-name>
kubectl logs -n kubesphere-system deploy/ks-controller-manager \
  | grep -E "helmrepo-controller|<repo-name>"
```

Common checks:

- Confirm `.spec.url` has a reachable `index.yaml`.
- Confirm credentials, CA, cert/key, and `insecureSkipTLSVerify` when using private HTTPS repositories.
- If apps disappeared after sync, check whether the chart was removed from the upstream index; the controller deletes apps no longer present for that repo.
- If sync loops, check whether the workspace label points to a deleted `WorkspaceTemplate`; the controller deletes workspace repos for deleted workspaces.

## Uploaded App Workflow

Uploaded apps are stored as `Repo=upload` and start in review state `draft`. Helm charts and YAML packages share the same API path; `appType` distinguishes `helm` and `yaml`.

Validation-only upload:

Authenticate through KubeSphere first. Prefer the `kubesphere-core` `ks_api.py` helper for JSON KAPIs because it handles login and cached tokens consistently with other KubeSphere skills. File uploads are multipart requests, so use `ks_api.py` to login and then curl with the cached token:

```bash
cd skills/kubesphere-core/scripts
export KUBESPHERE_HOST="http://<kubesphere-host>"
python ks_api.py --login --username admin --password <password>

TOKEN=$(python -c 'import json, os; print(json.load(open(os.path.expanduser("~/.kubesphere_token")))["token"])')

curl -X POST \
  "$KUBESPHERE_HOST/kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps?validate=true" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: multipart/form-data" \
  -F 'jsonData={"appType":"helm","workspace":"<workspace>"}' \
  -F "file=@./chart.tgz"
```

If the helper is unavailable, use curl with an explicit bearer token:

```bash
TOKEN=<kubesphere-access-token>
curl -X POST \
  "$KUBESPHERE_HOST/kapis/application.kubesphere.io/v2/workspaces/<workspace>/apps?validate=true" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: multipart/form-data" \
  -F 'jsonData={"appType":"helm","workspace":"<workspace>"}' \
  -F "file=@./chart.tgz"
```

After upload:

```bash
kubectl get applications.application.kubesphere.io \
  -l application.kubesphere.io/repo-name=upload,kubesphere.io/workspace=<workspace>

kubectl get applicationversions.application.kubesphere.io \
  -l application.kubesphere.io/repo-name=upload,application.kubesphere.io/app-id=<app>
```

Review states:

| State | Meaning |
|---|---|
| `draft` | Uploaded but not published. |
| `submitted` | Submitted for review. |
| `passed` | Review passed. |
| `active` | Published/visible. |
| `rejected` | Review rejected. |
| `suspended` | Temporarily hidden. |

Use app or version action routes to move review state. App activation requires at least one active or passed version.

## Release Workflow

An `ApplicationRelease` installs a selected `ApplicationVersion` into a target cluster and namespace.

Minimal Helm release object:

```yaml
apiVersion: application.kubesphere.io/v2
kind: ApplicationRelease
metadata:
  name: <release-name>
  labels:
    application.kubesphere.io/app-id: <app-id>
    application.kubesphere.io/appversion-id: <app-version-id>
    application.kubesphere.io/app-type: helm
    kubesphere.io/cluster: <cluster>
    kubesphere.io/namespace: <namespace>
    kubesphere.io/workspace: <workspace>
  annotations:
    kubesphere.io/creator: <username>
spec:
  appID: <app-id>
  appVersionID: <app-version-id>
  appType: helm
  values: <base64-or-api-provided-bytes>
```

Release states:

| State | Meaning |
|---|---|
| `creating` | First reconciliation started. |
| `created` | Helm/YAML executor Job was created. |
| `upgrading` | Spec changed and upgrade started. |
| `upgraded` | Upgrade Job was created. |
| `active` | Helm release deployed or YAML install completed. |
| `timeout` | Helm reported timeout; controller performs limited rechecks. |
| `deployFailed` | Executor Job failed or disappeared. |
| `failed` | Helm/YAML install, upgrade, or status verification failed. |
| `deleting` | Uninstall started. |
| `clusterDeleted` | Target cluster was deleted. |

Troubleshoot releases:

```bash
kubectl describe applicationrelease.application.kubesphere.io <release-name>

TARGET_NS=$(kubectl get applicationrelease.application.kubesphere.io <release-name> \
  -o jsonpath="{.metadata.labels['kubesphere.io/namespace']}")

kubectl -n "$TARGET_NS" get jobs \
  -l application.kubesphere.io/app-release-name=<release-name>

kubectl -n "$TARGET_NS" get pods \
  -l application.kubesphere.io/app-release-name=<release-name>
```

Always prefer the `application.kubesphere.io/app-release-name=<release-name>` label selector for executor Jobs and Pods. Do not use broad `kubectl get jobs -A | grep <release-name>` or `kubectl get pods -A | grep <release-name>` as the primary path; use grep only as a fallback when labels are missing or suspected to be wrong.

Then inspect the executor Job pod logs:

```bash
POD=$(kubectl -n "$TARGET_NS" get pods \
  -l application.kubesphere.io/app-release-name=<release-name> \
  -o jsonpath='{.items[0].metadata.name}')

kubectl -n "$TARGET_NS" logs "$POD" --all-containers
```

Common checks:

- Ensure `spec.appVersionID` exists and points to an `ApplicationVersion`.
- Ensure target cluster and namespace labels are correct; missing namespace defaults to `default`, missing cluster defaults to `host`.
- For Helm apps, check whether stored chart data can be loaded from S3 or the ConfigMap fallback.
- For YAML apps, verify `spec.values` contains valid YAML documents and the target cluster RESTMapper recognizes every GVR.
- For upgrade loops, compare `.status.specHash` with the current `.spec`; spec changes drive upgrades.
- For timeout, inspect annotation `application.kubesphere.io/timeout-recheck`; the controller only performs limited timeout rechecks.

## Categories

Categories are cluster-scoped resources. Application category is carried by `application.kubesphere.io/app-category-name`; uncategorized apps use `kubesphere-app-uncategorized`.

```bash
kubectl get categories.application.kubesphere.io
kubectl get applications.application.kubesphere.io \
  -l application.kubesphere.io/app-category-name=<category>
```

Do not delete a category until no applications reference it.

## Development Notes

When changing implementation:

- Prefer `application.kubesphere.io/v2` CRDs and KAPIs for new KubeSphere code.
- Keep backward compatibility in mind when touching older OpenPitrix extension paths under `/kapis/openpitrix.io/v2` and `/kapis/openpitrix.io/v2alpha1`.
- Preserve the object relationship: `Repo` owns synced apps, `Application` owns versions, and releases reference app/version through labels and spec fields.
- Status updates are subresource updates or merge patches; avoid normal spec updates for status-only changes.
- Uploaded package storage uses S3 when configured and falls back to ConfigMaps in `extension-openpitrix`.
- Keep review state transitions consistent with app and app-version action handlers.
