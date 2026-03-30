---
name: kubesphere-devops-tenant
description: Use when operating KubeSphere DevOps as a namespace-scoped tenant with limited permissions, without cluster-admin access, or when accessing DevOps through KubeSphere APIs only
---

# KubeSphere DevOps Tenant Operations

## Overview

This guide covers DevOps operations for **namespace-scoped tenants** who:
- Have admin/operator permissions within their DevOpsProject namespace(s)
- **Cannot** access `kubesphere-devops-system` (Jenkins secrets, tokens)
- **Cannot** call Jenkins APIs directly
- Must use **KubeSphere APIs** (`/kapis/devops.kubesphere.io/`) for all operations
- Use **KubeSphere authentication** (OAuth tokens), not Jenkins tokens

**Critical Distinction:** DevOps projects are **namespaces**, not DevOpsProject CRs. To list accessible DevOps projects:
```bash
# Correct - lists namespaces (DevOps projects) tenant can access
GET /clusters/{cluster}/kapis/devops.kubesphere.io/v1alpha3/workspaces/{workspace}/namespaces

# Wrong - requires cluster-admin, returns 403 for tenants
GET /clusters/{cluster}/apis/devops.kubesphere.io/v1alpha3/devopsprojects
```

## When to Use

- Operating as a project admin/operator (not cluster admin)
- Working within tenant namespace boundaries
- No access to Jenkins secrets in `kubesphere-devops-system`
- Need to trigger pipelines via KubeSphere API
- Building automation for namespace-scoped users
- Developing tenant-facing tooling

## Tenant vs Admin Permissions

| Capability | Tenant (Namespace) | Admin (Cluster) |
|------------|-------------------|-----------------|
| Access DevOpsProject | ✅ Own namespace(s) | ✅ All namespaces |
| Create/Edit Pipelines | ✅ In own namespace | ✅ Any namespace |
| View PipelineRuns | ✅ In own namespace | ✅ Any namespace |
| Access Jenkins Secret | ❌ No | ✅ `kubesphere-devops-system` |
| Direct Jenkins API | ❌ No | ✅ Full access |
| View Jenkins Console | ❌ No | ✅ Via NodePort |
| KubeSphere API | ✅ `/kapis/` | ✅ `/kapis/` |

## Authentication

Tenants authenticate via KubeSphere's OAuth, not Jenkins. See [kubesphere-core](../../core/kubesphere-core/SKILL.md) for complete OAuth authentication details.

### Quick Reference

```bash
# Exchange credentials for OAuth token (see core skill for details)
export KUBESPHERE_API="https://kubesphere-api.example.com"
export USERNAME="tenant-user"
export PASSWORD="tenant-password"

# Get token
export API_TOKEN=$(curl -s -X POST "${KUBESPHERE_API}/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password&username=${USERNAME}&password=${PASSWORD}&client_id=kubesphere&client_secret=kubesphere" \
  | jq -r '.access_token')

# Use token
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/namespaces/demo-project/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

**Key Points:**
- OAuth token expires in 7200 seconds (2 hours)
- Use `client_id=kubesphere` and `client_secret=kubesphere`
- Token contains user's RBAC permissions

See [kubesphere-core](../../core/kubesphere-core/SKILL.md#authentication) for complete OAuth authentication details including token refresh and common use cases.

### Get KubeSphere API Token

Tenants authenticate via KubeSphere's OAuth, not Jenkins:

```bash
# Method 1: Using kubeconfig (if configured)
kubectl config view --raw -o jsonpath='{.users[?(@.name=="current-user")].user.token}'

# Method 2: Via KubeSphere OAuth API (Recommended)
export KUBESPHERE_URL="https://kubesphere-api.example.com"
export USERNAME="tenant-user"
export PASSWORD="tenant-password"

# Exchange credentials for token
TOKEN_RESPONSE=$(curl -s -X POST "${KUBESPHERE_URL}/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode "grant_type=password" \
  --data-urlencode "username=${USERNAME}" \
  --data-urlencode "password=${PASSWORD}" \
  --data-urlencode "client_id=kubesphere" \
  --data-urlencode "client_secret=kubesphere")

# Extract access token
ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')

# Token expires in 7200 seconds (2 hours)
echo "Token obtained: ${ACCESS_TOKEN:0:50}..."
```

### Using Token with API

```bash
export API_TOKEN="<your-kubesphere-token>"
export DEVOPS_PROJECT="demo-project"
export KUBESPHERE_API="https://kubesphere-api.example.com"

# Verify access
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

## Complete Working Example

Here's a verified workflow using tenant credentials (stoneshi / P@88w0rd):

### Step 1: Authenticate
```bash
export KUBESPHERE_API="http://kubesphere-apiserver.kubesphere-system.svc:80"
export USERNAME="stoneshi"
export PASSWORD='P@88w0rd'

# Get OAuth token
TOKEN_RESPONSE=$(curl -s -X POST "${KUBESPHERE_API}/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "username=${USERNAME}" \
  -d "password=${PASSWORD}" \
  -d "client_id=kubesphere" \
  -d "client_secret=kubesphere")

export API_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
echo "Authenticated as: $(curl -s ${KUBESPHERE_API}/kapis/iam.kubesphere.io/v1beta1/users/stoneshi -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.metadata.name')"
```

### Step 2: Access Workspace Resources
```bash
# Verify workspace access (returns "stone")
curl -s "${KUBESPHERE_API}/kapis/tenant.kubesphere.io/v1beta1/workspaces/stone" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.metadata.name'

# Try accessing other workspace (returns 403 Forbidden - correct tenant isolation)
curl -s "${KUBESPHERE_API}/kapis/tenant.kubesphere.io/v1beta1/workspaces/demo" \
  -H "Authorization: Bearer ${API_TOKEN}"
# Output: {"message":"workspaces.tenant.kubesphere.io \"demo\" is forbidden..."}
```

### Step 3: Create and List Pipelines
```bash
export DEVOPS_PROJECT="stone-devops"  # Must be in "stone" workspace

# List pipelines in tenant namespace
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.items[] | "✓ " + .metadata.name'

# Create pipeline via kubectl (as tenant with namespace permissions)
cat <<EOF | kubectl apply -f -
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  name: stone-tenant-pipeline
  namespace: stone-devops
spec:
  type: pipeline
  pipeline:
    name: stone-tenant-pipeline
    description: "Test pipeline for tenant verification"
    jenkinsfile: |
      pipeline {
        agent { label "base" }
        stages {
          stage("Test") {
            steps {
              sh "echo 'Hello from tenant pipeline'"
            }
          }
        }
      }
EOF

# Verify via API
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/stone-tenant-pipeline" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '{name: .metadata.name, type: .spec.type}'
```

### Step 4: Trigger and Monitor Run
```bash
export PIPELINE_NAME="stone-tenant-pipeline"

# Trigger run
curl -s -X POST "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{}' | jq -r '{runId: .id, state: .state}'

# List runs (Blue Ocean format)
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq '.items[] | {id: .id, state: .state, result: .result}'

# Check specific run status
export RUN_ID="1"
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '{state: .state, result: .result, duration: .durationInMillis}'
# Output: {"state":"FINISHED","result":"SUCCESS","duration":15110}
```

### Step 5: Get Logs
```bash
# Get console log (tenant accessible, no Jenkins token needed)
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}/log" \
  -H "Authorization: Bearer ${API_TOKEN}" | tail -20

# Expected output includes:
# + echo Hello from tenant pipeline
# Hello from tenant pipeline
# Finished: SUCCESS
```

### Key Findings

| Aspect | Tenant Behavior |
|--------|-----------------|
| **Authentication** | OAuth with client_id/client_secret = "kubesphere" |
| **Token Expiry** | 7200 seconds (2 hours) |
| **API Version** | v1alpha3 for pipelines, v1alpha2 for runs |
| **Response Format** | Blue Ocean JSON (not Kubernetes resources) |
| **Status Fields** | `.state` (QUEUED/RUNNING/FINISHED), `.result` (SUCCESS/FAILURE) |
| **Namespace Isolation** | 403 Forbidden for other workspaces |
| **Logs Access** | ✅ Available via KubeSphere API |
| **Artifacts** | ✅ Available via `/artifacts` endpoint |

## Pipeline Operations

### List Pipelines (Tenant View)

```bash
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/search?q=type:pipeline" \
  -H "Authorization: Bearer ${API_TOKEN}"

# Or list in specific namespace
# Via API (v1alpha3 for pipelines)
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.items[].metadata.name'
  -H "Authorization: Bearer ${API_TOKEN}" | jq '.items[].metadata.name'
```

### Get Pipeline Details

```bash
export PIPELINE_NAME="my-pipeline"

curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq .
```

### Create Pipeline (Tenant)

**Regular Pipeline:**
```bash
curl -s -X POST "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "devops.kubesphere.io/v1alpha3",
    "kind": "Pipeline",
    "metadata": {
      "name": "my-tenant-pipeline",
      "namespace": "'${DEVOPS_PROJECT}'"
    },
    "spec": {
      "type": "pipeline",
      "pipeline": {
        "name": "my-tenant-pipeline",
        "description": "Pipeline created by tenant",
        "jenkinsfile": "pipeline {\n  agent { label \"base\" }\n  stages {\n    stage(\"Build\") {\n      steps {\n        sh \"echo Building...\"\n      }\n    }\n  }\n}"
      }
    }
  }'
```

**Multi-Branch Pipeline:**
```bash
curl -s -X POST "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "devops.kubesphere.io/v1alpha3",
    "kind": "Pipeline",
    "metadata": {
      "name": "my-multibranch-pipeline",
      "namespace": "'${DEVOPS_PROJECT}'"
    },
    "spec": {
      "type": "multi-branch-pipeline",
      "multi_branch_pipeline": {
        "name": "my-multibranch-pipeline",
        "description": "Multi-branch pipeline from tenant",
        "source_type": "git",
        "git_source": {
          "url": "https://github.com/example/repo",
          "credential_id": "my-git-credential",
          "discover_branches": true,
          "discover_tags": false
        },
        "script_path": "Jenkinsfile"
      }
    }
  }'
```

### Create Multi-Branch Pipeline from Private Repository

> ⚠️ **CRITICAL: Always Check Repository Type First**
> 
> Before creating any multi-branch pipeline, you **MUST** ask the user:
> > "Is this a private repository?"
> 
> **If YES (Private Repo):**
> 1. Ask if they want to use an existing credential or create a new one
> 2. Create a DevOps credential (`basic-auth` type with GitHub PAT) - see Step 1 below
> 3. Reference the credential in `git_source.credential_id` when creating the pipeline
> 
> **If NO (Public Repo):**
> - Set `credential_id: ""` (empty string)
> 
> Never assume repository type - always confirm with the user first. Never use `GITHUB_` env vars directly in pipeline specs.

**Complete workflow for private GitHub repository:**

**Prerequisites:**
- GitHub Personal Access Token (PAT) with repo access
- DevOps project namespace (e.g., `devopstestc2nj7`)

**Step 1: Create Credential for GitHub Access**
```bash
export GITHUB_PAT="ghp_xxxxxxxxxxxxxxxxxxxx"
export TENANT_NAME="stone-ns-admin"

curl -s -X POST "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/devopstestc2nj7/credentials" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
      "name": "github-token",
      "namespace": "devopstestc2nj7",
      "annotations": {
        "credential.devops.kubesphere.io/type": "basic-auth"
      }
    },
    "stringData": {
      "username": "git",
      "password": "'${GITHUB_PAT}'"
    },
    "type": "credential.devops.kubesphere.io/basic-auth"
  }'
```

**Step 2: Create GitRepository**
```bash
curl -s -X POST "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/devopstestc2nj7/gitrepositories" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "devops.kubesphere.io/v1alpha3",
    "kind": "GitRepository",
    "metadata": {
      "name": "my-private-repo",
      "namespace": "devopstestc2nj7"
    },
    "spec": {
      "url": "https://github.com/stoneshi-yunify/jenkinsfiles.git",
      "provider": "github",
      "secret": {
        "name": "github-token",
        "namespace": "devopstestc2nj7"
      },
      "description": "Private repository with Jenkinsfile"
    }
  }'
```

**Step 3: Create Multi-Branch Pipeline**
```bash
curl -s -X POST "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/devopstestc2nj7/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "devops.kubesphere.io/v1alpha3",
    "kind": "Pipeline",
    "metadata": {
      "name": "echo-pipeline",
      "namespace": "devopstestc2nj7",
      "annotations": {
        "kubesphere.io/creator": "'${TENANT_NAME}'"
      }
    },
    "spec": {
      "type": "multi-branch-pipeline",
      "multi_branch_pipeline": {
        "name": "echo-pipeline",
        "description": "Multi-branch pipeline from private repo",
        "source_type": "git",
        "git_source": {
          "url": "https://github.com/stoneshi-yunify/jenkinsfiles.git",
          "credential_id": "github-token",
          "discover_branches": true,
          "discover_tags": false
        },
        "script_path": "echo/Jenkinsfile"
      }
    }
  }'
```

**Key Points:**
- GitRepository requires `spec.provider` (e.g., `github`) and `spec.secret` fields
- Pipeline **MUST** have `kubesphere.io/creator` annotation when created by tenant
- Multi-branch pipelines auto-discover branches from the repository

## Pipeline Runs (The Tenant Way)

> ⚠️ **API Version Notice**: The `/kapis/devops.kubesphere.io/v1alpha2/` APIs are deprecated. Always prefer `v1alpha3` APIs when available.

### Trigger a Pipeline Run (Multi-Branch)

**For Multi-Branch Pipelines - Three-Step Procedure:**

**Step 1: List Available Branches**
```bash
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches?filter=origin&page=1&limit=10" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" | jq -r '.items[] | "- Branch: \(.name) | Latest: \(.latestRun.id // "N/A") | Status: \(.latestRun.result // "N/A")"'
```

**Step 2: Ask User Which Branch**
> "Which branch would you like to build?"

**Step 3: Trigger Build with Branch Parameter**
```bash
export BRANCH="main"  # User's selection

curl -s -X POST "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/pipelineruns?branch=${BRANCH}" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"parameters":[]}' | jq -r '.metadata.name'
```

**Key Points:**
- Use `v1alpha3` endpoint with `?branch=${BRANCH}` query parameter
- Returns Kubernetes PipelineRun resource (not Blue Ocean format)
- For multi-branch pipelines, the branch parameter is required

### Trigger Repository Scanning (Multi-Branch)

> **Exception to v1alpha3 rule**: Repository scanning uses **v1alpha2** API. This endpoint is not available in v1alpha3.

**When to use:**
- Force immediate repository re-scan to discover new branches
- Troubleshoot branch detection issues
- Manually trigger branch indexing after credential changes

**Step 1: Trigger Scan (v1alpha2)**
```bash
curl -X POST "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/scan" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Step 2: Fetch Scanning Log (v1alpha2)**
```bash
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/consolelog" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

**Example scanning output:**
```
Started by user stone-ns-admin
Starting branch indexing...
 > git ls-remote --symref -- https://github.com/org/repo.git
Fetching & pruning origin...
Checking branches:
  Checking branch main ✓
      'Jenkinsfile' found
    Met criteria
  Checking branch feature-branch
      'Jenkinsfile' found
    Met criteria
  Checking branch old-branch
      'Jenkinsfile' not found
    Does not meet criteria
Processed 3 branches
Finished branch indexing. Indexing took 3 sec
Finished: SUCCESS
```

**Via PipelineRun CR (kubectl) - Alternative:**
```bash
cat <<EOF | kubectl apply -f -
apiVersion: devops.kubesphere.io/v1alpha3
kind: PipelineRun
metadata:
  name: my-run-$(date +%s)
  namespace: ${DEVOPS_PROJECT}
spec:
  pipelineRef:
    name: ${PIPELINE_NAME}
  scm:
    refName: "main"    # Branch name for multi-branch pipelines
    refType: "branch"
EOF
```

### List Pipeline Runs

```bash
# Via kubectl (preferred - returns Kubernetes PipelineRun resources)
kubectl get pipelineruns -n ${DEVOPS_PROJECT} --sort-by=.metadata.creationTimestamp

# Via API (v1alpha3 - returns Kubernetes resources)
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelineruns?labelSelector=devops.kubesphere.io/pipeline=${PIPELINE_NAME}" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq '.items[] | {name: .metadata.name, phase: .status.phase, creationTime: .metadata.creationTimestamp}'
```

### Get Run Status

```bash
# Via kubectl (preferred)
kubectl get pipelinerun ${RUN_NAME} -n ${DEVOPS_PROJECT} -o jsonpath='{.status.phase}'

# Via API (v1alpha3)
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelineruns/${RUN_NAME}" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '{name: .metadata.name, phase: .status.phase, startTime: .status.startTime, completionTime: .status.completionTime}'
```

### Deprecated v1alpha2 APIs

> ⚠️ **Deprecated**: These v1alpha2 endpoints return Blue Ocean format and are deprecated. Use v1alpha3 APIs shown above.

```bash
# List runs (v1alpha2 - deprecated)
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq '.items[] | {id: .id, state: .state, result: .result}'

# Get run status (v1alpha2 - deprecated)
export RUN_ID="1"
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq '{id: .id, state: .state, result: .result, duration: .durationInMillis}'
```

```bash
# Get concise status (Blue Ocean format fields)
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '{id: .id, state: .state, result: .result, startTime: .startTime, duration: .durationInMillis}'

# Example output:
# {
#   "id": "1",
#   "state": "FINISHED",
#   "result": "SUCCESS",
#   "startTime": "2026-03-19T02:50:12.747+0000",
#   "duration": 15110
# }

# Watch for completion
while true; do
  STATUS=$(curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}" \
    -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.state')
  echo "State: $STATUS"
  [[ "$STATUS" == "FINISHED" ]] && break
  sleep 5
done
```

```bash
# Get concise status
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '{state: .status.phase, result: .status.conditions[0].reason, startTime: .status.startTime, completionTime: .status.completionTime}'

# Watch for completion
while true; do
  STATUS=$(curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}" \
    -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.status.phase')
  echo "Status: $STATUS"
  [[ "$STATUS" == "Succeeded" || "$STATUS" == "Failed" ]] && break
  sleep 5
done
```

## Logs and Artifacts (Tenant Access)

### Get Console Log

**Tenant Method (via KubeSphere API):**
```bash
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}/log" \
  -H "Authorization: Bearer ${API_TOKEN}"

# Or with kubectl
kubectl get pipelinerun ${RUN_ID} -n ${DEVOPS_PROJECT} -o jsonpath='{.status.log}' 2>/dev/null || echo "Logs via API only"
```

**Note:** Console logs may not be available immediately. Poll until ready:
```bash
while ! curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}/log" \
  -H "Authorization: Bearer ${API_TOKEN}" | grep -q "Finished:"; do
  echo "Waiting for logs..."
  sleep 5
done
echo "Logs ready!"
```

### List Artifacts

```bash
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}/artifacts" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq '.[] | {name: .name, path: .path, size: .size}'
```

### Download Artifacts

**Download via KubeSphere API:**
```bash
export ARTIFACT_NAME="service"
export ARTIFACT_PATH="service"

# Get artifact download URL
ARTIFACT_URL=$(curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}/artifacts" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r ".[] | select(.name==\"${ARTIFACT_NAME}\") | .url")

# Download artifact
curl -s "${KUBESPHERE_API}${ARTIFACT_URL}" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -o "/tmp/${ARTIFACT_NAME}"

# Verify
ls -lh "/tmp/${ARTIFACT_NAME}"
file "/tmp/${ARTIFACT_NAME}"
```

**Alternative: Via kubectl with exec (if artifact is in workspace):**
```bash
# Find the agent pod (if still running)
AGENT_POD=$(kubectl get pods -n kubesphere-devops-worker -l jenkins/label-digest -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

# Copy artifact (if pod exists)
if [ -n "$AGENT_POD" ]; then
  kubectl cp ${AGENT_POD}:/home/jenkins/agent/workspace/${PIPELINE_NAME}/${ARTIFACT_NAME} /tmp/${ARTIFACT_NAME}
fi
```

## Managing Credentials

### List Credentials

```bash
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/credentials" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq '.items[].metadata.name'
```

### Create Credential

**SSH Key:**
```bash
curl -s -X POST "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/credentials" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
      "name": "my-ssh-key",
      "namespace": "'${DEVOPS_PROJECT}'",
      "annotations": {
        "kubesphere.io/creator": "tenant-user",
        "kubesphere.io/description": "SSH key for Git"
      }
    },
    "type": "credential.devops.kubesphere.io/ssh",
    "stringData": {
      "username": "git",
      "privateKey": "'$(cat ~/.ssh/id_rsa | sed 's/$/\\n/g' | tr -d '\n')'"
    }
  }'
```

**Username/Password:**
```bash
curl -s -X POST "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/credentials" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
      "name": "my-basic-auth",
      "namespace": "'${DEVOPS_PROJECT}'"
    },
    "type": "credential.devops.kubesphere.io/basic-auth",
    "stringData": {
      "username": "myuser",
      "password": "mypassword"
    }
  }'
```

## Multi-Cluster Operations

KubeSphere supports managing DevOps resources across multiple clusters. Use the `/clusters/{cluster-name}/` prefix to forward API requests to specific member clusters.

### DevOps Projects are Namespaces

**Important:** In KubeSphere DevOps, a "DevOps project" is actually a Kubernetes namespace with the `devops.kubesphere.io/managed=true` label. The DevOpsProject CR is a wrapper resource, but when listing accessible DevOps projects for a tenant, you query **namespaces**, not DevOpsProject CRs.

**Correct API for listing tenant-accessible DevOps projects:**
```bash
# List DevOps project namespaces (NOT devopsprojects CRs)
GET /clusters/{cluster}/kapis/devops.kubesphere.io/v1alpha3/workspaces/{workspace}/namespaces
```

This endpoint returns namespaces that:
1. Have the `devops.kubesphere.io/managed=true` label
2. Have the `kubesphere.io/workspace={workspace}` label
3. Are accessible to the authenticated tenant

### DevOps Project Naming Convention

When users refer to DevOps projects, they may use either a **shortname** or **fullname**:

| Name Type | Example | Source | Description |
|-----------|---------|--------|-------------|
| **Shortname** | `devopstest` | DevOpsProject CR `.metadata.generateName` | User-friendly display name |
| **Fullname** | `devopstestc2nj7` | DevOpsProject CR `.metadata.name` | Actual namespace name |

**Key Points:**
- The **fullname** is the actual Kubernetes namespace name that you use in API calls
- The fullname = DevOpsProject CR's `.metadata.name` = Namespace's `.metadata.name`
- The shortname comes from `.metadata.generateName` and is used for display purposes

**Resolving Ambiguity:**
When a user provides a name that could match multiple projects:

```bash
# First, get all accessible DevOps project namespaces
NAMESPACES=$(curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/workspaces/stone2/namespaces" \
  -H "Authorization: Bearer ${API_TOKEN}")

# Example: User says "devopstest" which could match:
# - Fullname: devopstestc2nj7 (from generateName "devopstest")
# - Fullname: devopstestxyz12 (from generateName "devopstest")
# - Fullname: my-devopstest (different project)

# Check for matches
echo "$NAMESPACES" | jq -r '.items[].metadata.name' | grep "devopstest"
# Output might show:
# devopstestc2nj7
# devopstestxyz12

# If multiple matches found, ask user to confirm:
# "Multiple DevOps projects match 'devopstest':
#  1. devopstestc2nj7
#  2. devopstestxyz12
#  Which one do you want to use?"
```

**Best Practice:**
1. When user provides a name, check if it matches any fullname (namespace name) exactly
2. If exact match found → use that namespace
3. If no exact match → check if it matches any shortname (generateName prefix)
4. If multiple matches → **ask user to confirm** before proceeding

### API Path Patterns

| Endpoint Type | Path Pattern | Returns | Use Case | Tenant Access |
|--------------|--------------|---------|----------|---------------|
| **KubeSphere API (workspace-scoped)** | `/clusters/{cluster}/kapis/devops.kubesphere.io/v1alpha3/workspaces/{workspace}/namespaces` | **Namespaces** (DevOps projects) | List DevOps project namespaces tenant can access | ✅ **Tenant accessible** |
| **KubeSphere API (namespace-scoped)** | `/clusters/{cluster}/kapis/devops.kubesphere.io/v1alpha3/namespaces/{namespace}/pipelines` | Pipelines | Pipeline operations | ✅ **Tenant accessible** |
| **Kubernetes API (cluster-scoped)** | `/clusters/{cluster}/apis/devops.kubesphere.io/v1alpha3/devopsprojects` | DevOpsProject CRs | Direct CR access | ❌ **Admin only (403)** |

**Key Insight:** The `/kapis/` endpoints enforce workspace-level RBAC and work for tenants. The `/apis/` endpoints require cluster-scoped permissions and will return 403 for tenants. When listing DevOps projects a tenant can access, use the `/namespaces` endpoint, not `/devopsprojects`.

### List DevOpsProjects Across Clusters

```bash
export KUBESPHERE_API="http://kubesphere-apiserver:80"
export USERNAME="stone-ns-admin"
export PASSWORD="P@88w0rd"
export WORKSPACE="stone"

# Get OAuth token
TOKEN=$(curl -s -X POST -H 'Content-Type: application/x-www-form-urlencoded' \
  "${KUBESPHERE_API}/oauth/token" \
  --data-urlencode 'grant_type=password' \
  --data-urlencode "username=${USERNAME}" \
  --data-urlencode "password=${PASSWORD}" \
  --data-urlencode 'client_id=kubesphere' \
  --data-urlencode 'client_secret=kubesphere' | jq -r '.access_token')

# List DevOpsProjects on Host Cluster
echo "=== Host Cluster ==="
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/workspaces/${WORKSPACE}/namespaces" \
  -H "Authorization: Bearer ${TOKEN}" | jq -r '.items[] | "\(.metadata.name) (\(.metadata.creationTimestamp))"'

# List DevOpsProjects on Member-1 Cluster
echo "=== Member-1 Cluster ==="
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/workspaces/${WORKSPACE}/namespaces" \
  -H "Authorization: Bearer ${TOKEN}" | jq -r '.items[] | "\(.metadata.name) (\(.metadata.creationTimestamp))"'
```

**Example Output:**
```
=== Host Cluster ===
(No output - no DevOpsProjects in workspace 'stone' on host)

=== Member-1 Cluster ===
stonedev154cht (2026-03-18T06:46:04Z)
```

### Why Tenant Can't Use /apis/ Endpoints

```bash
# ❌ This will return 403 Forbidden for tenants
curl -s "${KUBESPHERE_API}/clusters/member-1/apis/devops.kubesphere.io/v1alpha3/devopsprojects" \
  -H "Authorization: Bearer ${TOKEN}"

# Output:
# {
#   "kind": "Status",
#   "apiVersion": "v1",
#   "status": "Failure",
#   "message": "devopsprojects.devops.kubesphere.io is forbidden: User \"stone-ns-admin\" cannot list resource \"devopsprojects\" in API group \"devops.kubesphere.io\" at the cluster scope",
#   "reason": "Forbidden",
#   "code": 403
# }

# ✅ This works because /kapis/ with workspace scope enforces tenant RBAC
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/workspaces/stone/namespaces" \
  -H "Authorization: Bearer ${TOKEN}"

# Output:
# {"items":[{"kind":"DevOpsProject","apiVersion":"devops.kubesphere.io/v1alpha3",...}]}
```

### Complete Multi-Cluster Query Script

```bash
#!/bin/bash

export KUBESPHERE_API="http://kubesphere-apiserver:80"
export USERNAME="stone-ns-admin"
export PASSWORD="P@88w0rd"
export WORKSPACE="stone"

# Get token
TOKEN=$(curl -s -X POST -H 'Content-Type: application/x-www-form-urlencoded' \
  "${KUBESPHERE_API}/oauth/token" \
  --data-urlencode 'grant_type=password' \
  --data-urlencode "username=${USERNAME}" \
  --data-urlencode "password=${PASSWORD}" \
  --data-urlencode 'client_id=kubesphere' \
  --data-urlencode 'client_secret=kubesphere' | jq -r '.access_token')

# Get list of clusters (requires admin token or cluster list permission)
# For tenants, typically hardcode the clusters they have access to
CLUSTERS=("host" "member-1")

echo "=== DevOpsProjects in Workspace '${WORKSPACE}' Across All Clusters ==="
for CLUSTER in "${CLUSTERS[@]}"; do
  echo -e "\n## Cluster: ${CLUSTER}"
  
  # Use /kapis/ endpoint with workspace scope
  ENDPOINT="${KUBESPHERE_API}"
  if [ "${CLUSTER}" != "host" ]; then
    ENDPOINT="${ENDPOINT}/clusters/${CLUSTER}"
  fi
  
  PROJECTS=$(curl -s "${ENDPOINT}/kapis/devops.kubesphere.io/v1alpha3/workspaces/${WORKSPACE}/namespaces" \
    -H "Authorization: Bearer ${TOKEN}")
  
  # Check if response contains items
  COUNT=$(echo "$PROJECTS" | jq '.items | length')
  
  if [ "$COUNT" -gt 0 ]; then
    echo "$PROJECTS" | jq -r '.items[] | "  - \(.metadata.name) (Created: \(.metadata.creationTimestamp), Status: \(.metadata.annotations."devopsproject.devops.kubesphere.io/syncstatus" // "N/A"))"'
  else
    echo "  No DevOpsProjects found"
  fi
done
```

## Workspace-Scoped API Operations

### Query DevOps Projects (Namespaces) by Workspace

Tenants can query DevOps projects (which are namespaces) within their authorized workspaces:

```bash
# List DevOps project namespaces in specific workspace
# Note: Returns namespaces with devops.kubesphere.io/managed=true label
# Returns empty if tenant doesn't have workspace access
curl -s "${KUBESPHERE_API}/clusters/host/kapis/devops.kubesphere.io/v1alpha3/workspaces/stone/namespaces?sortBy=createTime&limit=10" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.items[].metadata.name'

# Example output:
# stone-devops

# Query different workspace (returns 0 items if no access)
curl -s "${KUBESPHERE_API}/clusters/host/kapis/devops.kubesphere.io/v1alpha3/workspaces/demo/namespaces?sortBy=createTime&limit=10" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq '.totalItems'

# Example output:
# 0
```

**Important Distinction:**
- **DevOps Project** = A namespace with `devops.kubesphere.io/managed=true` label
- **DevOpsProject CR** = A Kubernetes custom resource that wraps the namespace
- To list projects a tenant can access → Use `/namespaces` endpoint
- The `/namespaces` endpoint filters by namespace label `kubesphere.io/workspace`. If the namespace label doesn't match the workspace, it won't be returned even if the DevOpsProject CR has the correct label.

### Verify Workspace Access

```bash
# Check accessible workspaces
curl -s "${KUBESPHERE_API}/kapis/tenant.kubesphere.io/v1beta1/workspaces" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.items[].metadata.name'

# Verify specific workspace
curl -s "${KUBESPHERE_API}/kapis/tenant.kubesphere.io/v1beta1/workspaces/stone" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.metadata.name'
```

## GitOps Application Deployment

### Create GitOps Application via API

Tenants can deploy applications using KubeSphere GitOps without accessing the ArgoCD namespace:

```bash
# Create GitOps Application
curl -s -X POST "${KUBESPHERE_API}/kapis/gitops.kubesphere.io/v1alpha1/namespaces/demo-project/applications" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "gitops.kubesphere.io/v1alpha1",
    "kind": "Application",
    "metadata": {
      "name": "guestbook",
      "namespace": "demo-project"
    },
    "spec": {
      "argoApp": {
        "spec": {
          "project": "default",
          "source": {
            "repoURL": "https://github.com/stoneshi-yunify/argocd-example-apps",
            "targetRevision": "HEAD",
            "path": "guestbook"
          },
          "destination": {
            "server": "https://kubernetes.default.svc",
            "namespace": "demo-project"
          },
          "syncPolicy": {
            "automated": {
              "prune": true,
              "selfHeal": true
            },
            "syncOptions": [
              "CreateNamespace=true"
            ]
          }
        }
      }
    }
  }' | jq -r '.metadata.name'

# Expected output: guestbook
```

### How It Works

1. **Tenant creates** `Application` (gitops.kubesphere.io/v1alpha1) in their namespace
2. **KubeSphere automatically** creates corresponding ArgoCD Application in `argocd` namespace
3. **ArgoCD controller** syncs the application to tenant's namespace
4. **Tenant cannot** access ArgoCD namespace directly - all operations via KubeSphere API

### Verify Application Deployment

**Method 1: Check Status Labels (Recommended for Tenants)**

The Application resource includes status labels that indicate the current health and sync status:

```bash
# Get Application and check status labels
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/gitops.kubesphere.io/v1alpha1/namespaces/demo-project/applications/guestbook" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '{
  name: .metadata.name,
  health: .metadata.labels["gitops.kubesphere.io/health-status"],
  sync: .metadata.labels["gitops.kubesphere.io/sync-status"],
  argocdApp: .metadata.labels["gitops.kubesphere.io/argocd-application"]
}'

# Expected output when synced and healthy:
# {
#   "name": "guestbook",
#   "health": "Healthy",
#   "sync": "Synced",
#   "argocdApp": "guestbook"
# }
```

**Status Values:**

| Label | Values | Description |
|-------|--------|-------------|
| `gitops.kubesphere.io/health-status` | Healthy, Progressing, Degraded, Missing, Unknown | Resource health state |
| `gitops.kubesphere.io/sync-status` | Synced, OutOfSync | Git repository sync state |

**Method 2: Check Detailed Status in .status.argoApp**

For more detailed information, parse the `.status.argoApp` field (JSON string):

```bash
# Get detailed sync and health information
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/gitops.kubesphere.io/v1alpha1/namespaces/demo-project/applications/guestbook" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.status.argoApp' | jq -r '{
  syncStatus: .sync.status,
  healthStatus: .health.status,
  revision: .sync.revision,
  resources: [.resources[] | {kind: .kind, name: .name, status: .status, health: .health.status}],
  images: .summary.images
}'

# Expected output:
# {
#   "syncStatus": "Synced",
#   "healthStatus": "Healthy",
#   "revision": "f946a1c393d50a460cc44944a476971fe13961f4",
#   "resources": [
#     {"kind": "Service", "name": "guestbook-ui", "status": "Synced", "health": "Healthy"},
#     {"kind": "Deployment", "name": "guestbook-ui", "status": "Synced", "health": "Healthy"}
#   ],
#   "images": ["gcr.io/google-samples/gb-frontend:v5"]
# }
```

**Method 3: Check Operation State**

For troubleshooting sync operations:

```bash
# Get operation state and sync result
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/gitops.kubesphere.io/v1alpha1/namespaces/demo-project/applications/guestbook" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.status.argoApp' | jq -r '.operationState | {
  phase: .phase,
  message: .message,
  startedAt: .startedAt,
  finishedAt: .finishedAt
}'

# Expected output on success:
# {
#   "phase": "Succeeded",
#   "message": "successfully synced (all tasks run)",
#   "startedAt": "2026-03-27T09:09:12Z",
#   "finishedAt": "2026-03-27T09:09:15Z"
# }
```

**Understanding Destination Cluster**

When `spec.argoApp.spec.destination.server` is `https://kubernetes.default.svc` and `destination.name` is empty or `in-cluster`, the Application deploys to the cluster specified in the API path:

| API Path | Destination Cluster |
|----------|-------------------|
| `/kapis/gitops.kubesphere.io/v1alpha1/namespaces/{ns}/applications` | Host cluster |
| `/clusters/member-1/kapis/gitops.kubesphere.io/v1alpha1/namespaces/{ns}/applications` | member-1 cluster |
| `/clusters/member-2/kapis/gitops.kubesphere.io/v1alpha1/namespaces/{ns}/applications` | member-2 cluster |

**Important for Tenants:**

Since tenants may not have permissions to directly query the destination namespace (due to RBAC), **always verify deployment via the Application status** rather than trying to access deployed resources directly:

```bash
# ✅ CORRECT: Check Application status (tenant has permissions)
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/gitops.kubesphere.io/v1alpha1/namespaces/demo-project/applications/guestbook" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.metadata.labels["gitops.kubesphere.io/sync-status"]'
# Output: "Synced"

# ❌ INCORRECT: Direct namespace access may fail for tenants
curl -s "${KUBESPHERE_API}/clusters/member-1/api/v1/namespaces/demo-project/pods" \
  -H "Authorization: Bearer ${API_TOKEN}"
# May return 403 Forbidden
```

### Tenant Limitations (Important)

| Action | Tenant Can | Notes |
|--------|-----------|-------|
| Create GitOps App | ✅ Yes | Via KubeSphere API |
| Modify ArgoCD Config | ❌ No | Cannot access `argocd` namespace |
| Add App Namespace to ArgoCD | ❌ No | Requires admin to update `application.namespaces` |
| View ArgoCD UI | ❌ No | No direct ArgoCD access |
| View Deployed Resources | ✅ Yes | In own namespace |

## Complete Tenant Workflow

### Step-by-Step: Build and Retrieve Artifacts as Tenant

```bash
#!/bin/bash
set -e

# Configuration
export KUBESPHERE_API="https://kubesphere-api.example.com"
export API_TOKEN="<tenant-token>"
export DEVOPS_PROJECT="demo-project"
export PIPELINE_NAME="my-tenant-pipeline"

# 1. List available pipelines
echo "=== Available Pipelines ==="
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.items[].metadata.name'

# 2. Trigger pipeline run
echo "=== Triggering Pipeline ==="
RUN_RESPONSE=$(curl -s -X POST "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"parameters": [{"name": "BRANCH", "value": "main"}]}')

RUN_ID=$(echo $RUN_RESPONSE | jq -r '.metadata.name')
echo "Run ID: $RUN_ID"

# 3. Wait for completion
echo "=== Waiting for Build ==="
while true; do
  STATUS=$(curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}" \
    -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.status.phase')
  echo "Status: $STATUS"
  [[ "$STATUS" == "Succeeded" || "$STATUS" == "Failed" ]] && break
  sleep 10
done

# 4. Get logs
echo "=== Console Log ==="
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}/log" \
  -H "Authorization: Bearer ${API_TOKEN}" | tail -50

# 5. Download artifacts
echo "=== Downloading Artifacts ==="
ARTIFACTS=$(curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/runs/${RUN_ID}/artifacts" \
  -H "Authorization: Bearer ${API_TOKEN}")

echo "$ARTIFACTS" | jq -c '.[]' | while read artifact; do
  NAME=$(echo $artifact | jq -r '.name')
  URL=$(echo $artifact | jq -r '.url')
  echo "Downloading: $NAME"
  curl -s "${KUBESPHERE_API}${URL}" -H "Authorization: Bearer ${API_TOKEN}" -o "/tmp/${NAME}"
  ls -lh "/tmp/${NAME}"
done

echo "=== Done ==="
```

## Tenant Limitations & Workarounds

| Limitation | Tenant Impact | Workaround |
|------------|---------------|------------|
| No Jenkins token | Cannot use Jenkins API directly | Use KubeSphere `/kapis/` endpoints |
| No kubesphere-devops-system access | Cannot view Jenkins master logs | View PipelineRun status via API |
| No agent pod access | Cannot exec into agents | Artifacts via API or pipeline steps |
| Limited logs | Logs may be truncated | Store logs in artifacts or external systems |
| No webhook management | Cannot configure webhooks directly | Use KubeSphere UI or request admin |

### Common Errors and Fixes

**Error: 403 Forbidden**
```bash
# Cause: Token expired or insufficient permissions
# Fix: Refresh token or check RBAC

curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}" -v 2>&1 | grep "HTTP/"
# Should be: HTTP/2 200
```

**Error: Resource not found**
```bash
# Cause: Wrong namespace or resource doesn't exist
# Fix: Verify namespace and resource names

kubectl get pipelines -n ${DEVOPS_PROJECT}
kubectl auth can-i get pipelines -n ${DEVOPS_PROJECT}
```

**Error: No logs available**
```bash
# Cause: Run not complete or logs not persisted
# Fix: Wait for completion, check if run succeeded

kubectl get pipelinerun ${RUN_ID} -n ${DEVOPS_PROJECT} -o jsonpath='{.status.phase}'
```


### Workspace API Returns Empty (Namespace Label Mismatch)

**Symptom:**
```bash
# Query workspace API returns 0 items even though DevOpsProject exists
curl -s "${KUBESPHERE_API}/clusters/host/kapis/devops.kubesphere.io/v1alpha3/workspaces/demo/namespaces" \
  -H "Authorization: Bearer ${API_TOKEN}"
# Output: {"items": null, "totalItems": 0}

# But direct query works
curl -s "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/namespaces/demo-project/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}"
# Output: Returns pipelines successfully
```

**Root Cause:**
The workspace-scoped API filters namespaces by the label `kubesphere.io/workspace`. If the **namespace** label doesn't match the workspace, it won't be returned, even if the DevOpsProject CR has the correct label.

**Check Labels:**
```bash
# Check DevOpsProject label (usually correct)
kubectl get devopsproject demo-project -o jsonpath='{.metadata.labels.kubesphere\.io/workspace}'
# Output: demo

# Check namespace label (may be empty or wrong)
kubectl get ns demo-project -o jsonpath='{.metadata.labels.kubesphere\.io/workspace}'
# Output: "" (EMPTY - this is the problem!)
```

**Fix (Admin Required):**
```bash
# Update namespace label to match workspace
kubectl label ns demo-project kubesphere.io/workspace=demo --overwrite

# Verify fix
kubectl get ns demo-project -o jsonpath='{.metadata.labels.kubesphere\.io/workspace}'
# Output: demo

# Now workspace API returns the namespace
curl -s "${KUBESPHERE_API}/clusters/host/kapis/devops.kubesphere.io/v1alpha3/workspaces/demo/namespaces" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq '.totalItems'
# Output: 1
```

**Why This Happens:**
- DevOpsProject CR and namespace are separate resources
- DevOpsProject controller should sync the workspace label to the namespace
- If the controller missed it or the label was removed, the API filtering breaks
- Workspace-scoped APIs use namespace labels, not DevOpsProject labels

## References

- [KubeSphere DevOps Overview](../kubesphere-devops-overview/SKILL.md)
- [KubeSphere DevOps Pipeline](../kubesphere-devops-pipeline/SKILL.md)
- [DevOps API Documentation](https://docs.kubesphere.io/)
- [RBAC in KubeSphere](https://docs.kubesphere.io/v4.1/05-access-control-and-account-management/)
