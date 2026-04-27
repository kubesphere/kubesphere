---
name: kubesphere-devops-pipeline
description: Use when creating, running, or managing CI/CD pipelines in KubeSphere DevOps, including pipeline API operations and run monitoring
---

# KubeSphere DevOps Pipeline Management

## Overview

Pipelines in KubeSphere DevOps are Kubernetes custom resources that integrate with Jenkins. KubeSphere uses a cloud-native, object-reconcile approach where Kubernetes resources are the source of truth.

## When to Use

- Creating or updating CI/CD pipelines
- Triggering pipeline runs
- Monitoring pipeline execution
- Retrieving pipeline logs and artifacts
- Troubleshooting failed pipeline runs

## Architecture Mapping

KubeSphere DevOps maps Kubernetes resources to Jenkins objects:

| KubeSphere Resource | K8s Resource | Jenkins Resource |
|---------------------|--------------|------------------|
| DevOpsProject | DevOpsProject CR + Namespace | Folder |
| Pipeline | Pipeline CR | WorkflowJob |
| PipelineRun | PipelineRun CR | Build Run |
| Workspace | Workspace CR | (authorization context) |

```
KubeSphere          Kubernetes                Jenkins
─────────────────────────────────────────────────────────────
Workspace demo
└── DevOpsProject   → demo-project NS          → Folder demo-project
    └── Pipeline    → Pipeline CR              → WorkflowJob
        └── Run     → PipelineRun CR           → Build #1
```

## Triggering Pipeline Runs (Recommended: Object-Reconcile)

> **CRITICAL: ALWAYS Check for Parameters First!**
> 
> Before triggering ANY pipeline (regular or multi-branch), you MUST check if the pipeline has parameters defined.
> Triggering a pipeline without required parameters will cause the build to fail or use incorrect defaults.
> 
> **For Multi-Branch Pipelines:** Query `/branches/{branch}` endpoint to get `.parameters` array
> **For Regular Pipelines:** Query the Pipeline CR and check `.spec.pipeline.jenkinsfile` for `parameters {}` directive

**Preferred Approach:** Create a `PipelineRun` custom resource. KubeSphere watches for these resources and triggers the corresponding Jenkins build.

### Create a PipelineRun

```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: PipelineRun
metadata:
  name: my-pipeline-run-001
  namespace: demo-project
spec:
  pipelineRef:
    name: my-pipeline
  parameters:
    - name: BRANCH
      value: "main"
```

Apply with kubectl:
```bash
kubectl apply -f pipelinerun.yaml
```

### Check PipelineRun Status

```bash
# List all runs
kubectl get pipelineruns -n demo-project

# Get specific run status
kubectl get pipelinerun my-pipeline-run-001 -n demo-project -o yaml

# Watch run progress
kubectl get pipelineruns -n demo-project -w
```

### PipelineRun Status Fields

| Field | Description |
|-------|-------------|
| `status.phase` | Current state (Pending, Running, Succeeded, Failed, Unknown) |
| `status.conditions` | Detailed conditions (Succeeded, Ready) |
| `status.completionTime` | When run finished |
| `status.startTime` | When run started |

### Delete a PipelineRun

```bash
kubectl delete pipelinerun my-pipeline-run-001 -n demo-project
```

## Working Pipeline Example

Here's a complete, working pipeline that builds a Go application:

```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  name: go-demo-pipeline
  namespace: demo-project
spec:
  type: pipeline
  pipeline:
    name: go-demo-pipeline
    description: "Build and test Go application"
    jenkinsfile: |
      pipeline {
        agent any

        stages {
          stage('Build, Test and Archive') {
            agent {
              kubernetes {
                yaml '''
                  apiVersion: v1
                  kind: Pod
                  spec:
                    containers:
                    - name: golang
                      image: golang:1.21
                      command: ["sleep"]
                      args: ["99d"]
                '''
              }
            }
            steps {
              container('golang') {
                sh '''
                  export GO111MODULE=on
                  git clone https://github.com/kubesphere-sigs/demo-go-http.git .
                  go mod download
                  go test ./... -v
                  go build -o service main.go
                '''
              }
              archiveArtifacts artifacts: 'service', followSymlinks: false
            }
          }
        }
      }
```

**Key Points:**
- Uses `agent { kubernetes { yaml ... } }` to define a custom pod with Go container
- The `archiveArtifacts` step must be in the same stage as the build (same workspace)
- Container name in `container('golang')` must match the container name in the YAML

## Pipeline Resource

```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  name: my-pipeline
  namespace: demo-project
spec:
  type: pipeline
  pipeline:
    name: my-pipeline
    description: "Build and deploy app"
    jenkinsfile: |-
      pipeline {
        agent any
        stages {
          stage('Build') {
            steps {
              sh 'make build'
            }
          }
        }
      }
```

### Tenant-Created Resources: Creator Annotation

When creating pipelines as a **tenant** (not cluster-admin), you **MUST** include the `kubesphere.io/creator` annotation to properly track ownership:

```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  name: my-pipeline
  namespace: demo-project
  annotations:
    kubesphere.io/creator: "stone-ns-admin"  # Required for tenant-created resources
spec:
  type: pipeline
  pipeline:
    name: my-pipeline
    description: "Pipeline created by tenant"
    jenkinsfile: |-
      pipeline {
        agent any
        stages {
          stage('Build') {
            steps {
              sh 'echo Building...'
            }
          }
        }
      }
```

**Why this matters:**
- KubeSphere uses this annotation for ownership tracking
- UI displays creator information
- Required for proper RBAC enforcement
- **CRITICAL: Always set this annotation when creating ANY pipeline via API (both regular and multi-branch)**

**Example API call with creator annotation for regular pipeline:**
```bash
curl -s -X POST "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/namespaces/demo-project/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "devops.kubesphere.io/v1alpha3",
    "kind": "Pipeline",
    "metadata": {
      "name": "my-pipeline",
      "namespace": "demo-project",
      "annotations": {
        "kubesphere.io/creator": "'${USERNAME}'"
      }
    },
    "spec": {
      "type": "pipeline",
      "pipeline": {
        "name": "my-pipeline",
        "description": "Tenant-created pipeline",
        "jenkinsfile": "pipeline { agent any; stages { stage(\"Build\") { steps { sh \"echo hello\" } } } }"
      }
    }
  }'
```

## Multi-Branch Pipeline

Multi-branch pipelines automatically discover branches from SCM and create jobs for each branch. The Jenkinsfile is loaded from the repository.

> ⚠️ **CRITICAL: Always Check Repository Type First**
> 
> Before creating a multi-branch pipeline, you **MUST** ask the user:
> > "Is this a private repository?"
> 
> **If YES (Private Repo):**
> 1. Ask if they want to use an existing credential or create a new one
> 2. Create a DevOps credential (`basic-auth` type with GitHub PAT) if needed
> 3. Reference the credential in `git_source.credential_id`
> 4. (Optional) Create a GitRepository CR for additional metadata
> 
> **If NO (Public Repo):**
> - Set `credential_id: ""` (empty string)
> 
> Never assume repository type - always confirm with the user first.

### Create Multi-Branch Pipeline

**Step 1: Check Repository Type**
- Ask user: "Is the repository private?"
- If yes, proceed with credential creation

**Step 2: Create Credential (For Private Repos Only)**

```bash
# Create GitHub credential
export GITHUB_PAT="ghp_xxxxxxxxxxxxxxxxxxxx"

curl -s -X POST "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/credentials" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
      "name": "github-token",
      "namespace": "'${DEVOPS_PROJECT}'",
      "annotations": {
        "kubesphere.io/creator": "'${USERNAME}'",
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

**Step 3a: For Public Repository:**
```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  name: demo-jenkinsfiles-go
  namespace: demo-project
  annotations:
    kubesphere.io/creator: "stone-ns-admin"  # Required for tenant-created resources
spec:
  type: multi-branch-pipeline
  multi_branch_pipeline:
    name: demo-jenkinsfiles-go
    description: "Multi-branch Go pipeline"
    source_type: git
    git_source:
      url: https://github.com/kubesphere/demo-jenkinsfiles
      credential_id: ""  # Empty for public repos
      discover_branches: true
      discover_tags: false
    script_path: go/Jenkinsfile  # Path to Jenkinsfile in repo
```

**Step 3b: For Private Repository (with credential):**
```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  name: private-repo-pipeline
  namespace: demo-project
  annotations:
    kubesphere.io/creator: "stone-ns-admin"
spec:
  type: multi-branch-pipeline
  multi_branch_pipeline:
    name: private-repo-pipeline
    description: "Pipeline for private repository"
    source_type: git
    git_source:
      url: https://github.com/org/private-repo.git
      credential_id: "github-token"  # Reference to DevOps credential
      discover_branches: true
      discover_tags: false
    script_path: Jenkinsfile
```

**Complete Flow for Private Repo:**
1. **Ask user** if repository is private (ALWAYS do this first)
2. **Create credential** (`basic-auth` type with GitHub PAT)
3. **(Optional) Create GitRepository** CR (with `provider` and `secret` fields)
4. **Create multi-branch pipeline** referencing the credential in `git_source.credential_id`

**Important:** Never use `GITHUB_` env vars directly in the pipeline spec. Always create proper DevOps credentials.

**SCM Source Types:**

| Type | Field | Use Case |
|------|-------|----------|
| Git | `git_source` | Generic Git repositories |
| GitHub | `github_source` | GitHub.com or GitHub Enterprise |
| GitLab | `gitlab_source` | GitLab.com or self-hosted GitLab |
| SVN | `svn_source` | Subversion repositories |

### Trigger Multi-Branch Pipeline Run

```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: PipelineRun
metadata:
  name: demo-jenkinsfiles-go-main-run
  namespace: demo-project
spec:
  pipelineRef:
    name: demo-jenkinsfiles-go
  scm:
    refName: main    # Branch name
    refType: branch  # or 'tag'
```

### Check Discovered Branches

**Via v1alpha3 API (preferred):**
```bash
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches?filter=origin&page=1&limit=10" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" | jq -r '.items[] | "- Branch: \(.name) | Latest Run: \(.latestRun.id // "N/A") | Status: \(.latestRun.result // "N/A")"'
```

**Via Jenkins (admin only):**
```bash
kubectl run curl-jenkins --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/job/demo-jenkinsfiles-go/api/json"
```

### Trigger Repository Scanning

> **Note**: Repository scanning uses **v1alpha2** API (not v1alpha3). This is an exception to the general rule of preferring v1alpha3.

**Step 1: Trigger Scan**
```bash
curl -X POST "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/scan" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Step 2: Fetch Scanning Log**
```bash
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/consolelog" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

**When to use:**
- Force immediate repository re-scan
- Discover new branches that aren't showing up
- Troubleshoot branch detection issues

**Example scanning log output:**
```
Started by user admin
Starting branch indexing...
 > git ls-remote --symref -- https://github.com/org/repo.git
Checking branches:
  Checking branch main ✓
      'Jenkinsfile' found
    Met criteria
  Checking branch feature-x ✓
      'Jenkinsfile' found
    Met criteria
  Checking branch old-branch
      'Jenkinsfile' not found
    Does not meet criteria
Processed 3 branches
Finished branch indexing. Indexing took 3 sec
Finished: SUCCESS
```

### Complete Private Repo Setup Example

**Scenario:** Create a multi-branch pipeline for a private GitHub repository with proper authentication.

**Step 1: Ask User About Repository Type**
```
Question: "Is https://github.com/stoneshi-yunify/jenkinsfiles a private repository?"
User Answer: "Yes"
```

**Step 2: Create GitHub Credential**
```bash
export KUBESPHERE_API="http://kubesphere-apiserver:80"
export API_TOKEN="<tenant-oauth-token>"
export DEVOPS_PROJECT="devopstestc2nj7"
export USERNAME="stone-ns-admin"
export GITHUB_PAT="ghp_xxxxxxxxxxxxxxxxxxxx"

curl -s -X POST "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/credentials" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
      "name": "github-token",
      "namespace": "'${DEVOPS_PROJECT}'",
      "annotations": {
        "kubesphere.io/creator": "'${USERNAME}'",
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

**Step 3: Create Multi-Branch Pipeline with Credential**
```bash
curl -s -X POST "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "devops.kubesphere.io/v1alpha3",
    "kind": "Pipeline",
    "metadata": {
      "name": "echo-jenkinsfile-pipeline",
      "namespace": "'${DEVOPS_PROJECT}'",
      "annotations": {
        "kubesphere.io/creator": "'${USERNAME}'"
      }
    },
    "spec": {
      "type": "multi-branch-pipeline",
      "multi_branch_pipeline": {
        "name": "echo-jenkinsfile-pipeline",
        "description": "Multi-branch pipeline with GitHub auth",
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

**Note:** The `kubesphere.io/creator` annotation is **required** for all pipelines (both regular and multi-branch). Without it, the pipeline may not appear correctly in the KubeSphere UI.

**Step 4: Verify Creation**
```bash
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/echo-jenkinsfile-pipeline" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '{
    name: .metadata.name,
    syncStatus: .metadata.annotations."pipeline.devops.kubesphere.io/syncstatus",
    credential: .spec.multi_branch_pipeline.git_source.credential_id
  }'
```

Expected output:
```json
{
  "name": "echo-jenkinsfile-pipeline",
  "syncStatus": "successful",
  "credential": "github-token"
}
```

---

### Complete Multi-Branch Workflow Example

**Scenario:** Trigger build, monitor, retrieve logs, and download artifacts.

**Step 1: Check Pipeline Exists**
```bash
kubectl get pipelines -n demo-project
# NAME                   TYPE                    AGE
# demo-jenkinsfiles-go   multi-branch-pipeline   12d
```

**Step 2: Trigger Build via Jenkins API**
```bash
# Get token
TOKEN=$(kubectl -n kubesphere-devops-system get secret devops-jenkins -o jsonpath='{.data.jenkins-admin-token}' | base64 -d)

# Trigger main branch build
kubectl run curl-trigger --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/job/demo-jenkinsfiles-go/job/main/build" \
  -X POST -w "\nHTTP Status: %{http_code}\n"

# Expected: HTTP Status: 201
```

**Step 3: Monitor Build Status**
```bash
# Check PipelineRun created
kubectl get pipelineruns -n demo-project --sort-by=.metadata.creationTimestamp | tail -3

# Example output:
# demo-jenkinsfiles-go-vf8p5       3     Succeeded   2m

# Get detailed status
# Or check via Jenkins API
kubectl run curl-status --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/job/demo-jenkinsfiles-go/job/main/3/api/json" \
  | grep -E '"result"|"building"|"duration"'
```

**Step 4: Retrieve Console Log**
```bash
# Get full console log from build #3
kubectl run curl-log --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/job/demo-jenkinsfiles-go/job/main/3/consoleText"

# Look for:
# [Pipeline] Start of Pipeline
# [Pipeline] { (Clone Repository)
# [Pipeline] { (Run Tests)
# === RUN   TestHello
# --- PASS: TestHello (0.00s)
# [Pipeline] End of Pipeline
# Finished: SUCCESS
```

**Step 5: Download Build Artifacts**

For binary artifacts, use pod-based download:

```bash
# Create download pod
kubectl run artifact-downloader --image=curlimages/curl -- sleep 300
kubectl wait --for=condition=Ready pod/artifact-downloader --timeout=60s

# Download artifact inside pod
TOKEN=$(kubectl -n kubesphere-devops-system get secret devops-jenkins -o jsonpath='{.data.jenkins-admin-token}' | base64 -d)
kubectl exec artifact-downloader -- sh -c \
  "curl -s -o /tmp/service 'http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/job/demo-jenkinsfiles-go/job/main/3/artifact/service'"

# Copy to local
kubectl cp artifact-downloader:/tmp/service /tmp/service

# Clean up
kubectl delete pod artifact-downloader --force
```

**Step 6: Verify Downloaded Artifact**
```bash
# Check file details
ls -lh /tmp/service
file /tmp/service

# Expected output:
# /tmp/service: ELF 64-bit LSB executable, x86-64...

# Make executable and test
chmod +x /tmp/service
/tmp/service --help
```

**Build Summary Example:**
```
Build #3 Summary:
- Status: SUCCESS
- Duration: ~54 seconds
- Test Results: 1/1 passed
- Artifact: service (8.0 MB Go binary)
- Stages: Checkout → Clone → Dependencies → Test → Build → Archive
```

### Multi-Branch vs Regular Pipeline

| Feature | Regular Pipeline | Multi-Branch Pipeline |
|---------|------------------|----------------------|
| Type | `pipeline` | `multi-branch-pipeline` |
| Jenkinsfile | Inline in CRD | From SCM (`script_path`) |
| Branches | Single | Auto-discovered |
| SCM Config | Manual checkout | Automatic |
| Trigger | PipelineRun/API | Branch indexing + webhooks |
| Get Parameters | From `.spec.pipeline.jenkinsfile` | From `/branches/{branch}` endpoint |

## Triggering Regular Pipeline Runs with Parameters (v1alpha3 API)

For regular pipelines, use this **two-step procedure** to handle parameters:

### Step 1: Get Pipeline Definition and Extract Parameters

Unlike multi-branch pipelines, regular pipelines don't have a dedicated `/parameters` endpoint. Instead, retrieve the Pipeline CR and extract parameters from the Jenkinsfile:

```bash
# Get the pipeline definition
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" | jq -r '.spec.pipeline.jenkinsfile'
```

**Look for the `parameters {}` directive in the Jenkinsfile:**

```groovy
pipeline {
  agent any
  stages {
    stage('Example') {
      steps {
        echo "Hello ${params.PERSON}"
      }
    }
  }
  
  parameters {
    string(name: 'PERSON', defaultValue: 'Mr Jenkins', description: 'Who should I say hello to?')
    booleanParam(name: 'TOGGLE', defaultValue: true, description: 'Toggle this value')
    choice(name: 'CHOICE', choices: ['One', 'Two', 'Three'], description: 'Pick something')
  }
}
```

**Parameter Types in Jenkinsfile:**

| Directive | Type | Example |
|-----------|------|---------|
| `string()` | Single-line text | `string(name: 'VAR', defaultValue: 'default', description: '...')` |
| `text()` | Multi-line text | `text(name: 'VAR', defaultValue: '', description: '...')` |
| `booleanParam()` | True/false | `booleanParam(name: 'FLAG', defaultValue: true, description: '...')` |
| `choice()` | Dropdown | `choice(name: 'ENV', choices: ['dev', 'prod'], description: '...')` |
| `password()` | Hidden value | `password(name: 'SECRET', defaultValue: '', description: '...')` |

### Step 2: Trigger Pipeline with Parameters

Use the `/pipelineruns` endpoint to create a new run with parameters:

```bash
# Trigger with parameters
curl -s -X POST "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/pipelineruns" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "parameters": [
      {"name": "PERSON", "value": "John Doe"},
      {"name": "TOGGLE", "value": "true"},
      {"name": "CHOICE", "value": "Two"}
    ]
  }' | jq -r '.metadata.name'
```

**Key Points:**
- No branch query parameter needed (regular pipelines have no branches)
- Parameters are passed in the request body as a JSON array
- Boolean values must be strings: `"true"` or `"false"`
- The response includes the created PipelineRun resource with your parameters

### Complete Example for Regular Pipelines

```bash
#!/bin/bash
export KUBESPHERE_API="http://kubesphere-apiserver:80"
export API_TOKEN="<tenant-oauth-token>"
export DEVOPS_PROJECT="devopstestc2nj7"
export PIPELINE_NAME="my-regular-pipeline"

# Step 1: Get pipeline definition and check for parameters
echo "Checking for parameters in pipeline..."
JENKINSFILE=$(curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}" \
  -H "Authorization: Bearer ${API_TOKEN}" | jq -r '.spec.pipeline.jenkinsfile')

# Check if parameters exist
if echo "$JENKINSFILE" | grep -q "parameters {"; then
  echo "Pipeline has parameters defined. Extracting..."
  # In practice, parse the Jenkinsfile to extract parameter names and defaults
  echo "$JENKINSFILE" | grep -E "(string|booleanParam|choice|password|text)\(name:"
else
  echo "No parameters found. Triggering without parameters."
fi

# Step 2: Trigger with parameters
echo "Triggering pipeline with parameters..."
curl -s -X POST "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/pipelineruns" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "parameters": [
      {"name": "PERSON", "value": "Test User"},
      {"name": "TOGGLE", "value": "true"},
      {"name": "CHOICE", "value": "Two"},
      {"name": "BIOGRAPHY", "value": "Testing API"},
      {"name": "PASSWORD", "value": "secret123"}
    ]
  }' | jq -r '.metadata.name'
```

## Triggering Multi-Branch Pipeline Runs (v1alpha3 API)

> ⚠️ **API Version Notice**: The `/kapis/devops.kubesphere.io/v1alpha2/` APIs are deprecated. Always prefer `v1alpha3` APIs when available.

For multi-branch pipelines, use this **three-step procedure** with v1alpha3 APIs:

### Step 1: List Available Branches

```bash
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches?filter=origin&page=1&limit=10" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" | jq -r '.items[] | "- Branch: \(.name) | Latest Run: \(.latestRun.id // "N/A") | Status: \(.latestRun.result // "N/A")"'
```

Example output:
```
- Branch: main | Latest Run: 2 | Status: SUCCESS
- Branch: stone | Latest Run: 1 | Status: SUCCESS
```

### Step 2: Ask User Which Branch to Build

Present the available branches to the user and ask:
> "Which branch would you like to build?"

### Step 3: Trigger Build on Selected Branch

```bash
export BRANCH="main"  # User's selection

curl -s -X POST "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/pipelineruns?branch=${BRANCH}" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"parameters":[]}' | jq -r '.metadata.name'
```

**Key Points:**
- Use `v1alpha3` endpoint (not v1alpha2)
- Pass branch as query parameter `?branch=${BRANCH}`
- Returns a PipelineRun resource (Kubernetes CRD format)
- The PipelineRun name is auto-generated with the pipeline name as prefix

### Triggering with Parameters

When a pipeline has parameters defined in the Jenkinsfile, you can provide values when triggering a build. Follow this workflow:

#### Step 1: Get Available Parameters

Query the branch to retrieve parameter definitions:

```bash
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches/${BRANCH}" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" | jq '.parameters'
```

**Example Response:**
```json
[
  {
    "description": "Who should I say hello to?",
    "name": "PERSON",
    "type": "StringParameterDefinition",
    "value": "",
    "defaultParameterValue": {
      "name": "PERSON",
      "value": "Mr Jenkins"
    }
  },
  {
    "description": "Toggle this value",
    "name": "TOGGLE",
    "type": "BooleanParameterDefinition",
    "value": "",
    "defaultParameterValue": {
      "name": "TOGGLE",
      "value": true
    }
  },
  {
    "description": "Pick something",
    "name": "CHOICE",
    "type": "ChoiceParameterDefinition",
    "value": "",
    "defaultParameterValue": {
      "name": "CHOICE",
      "value": "One"
    },
    "choices": ["One", "Two", "Three"]
  }
]
```

**Parameter Types:**

| Type | Description | Value Format |
|------|-------------|--------------|
| `StringParameterDefinition` | Single-line text | `"value": "text"` |
| `TextParameterDefinition` | Multi-line text | `"value": "multi\\nline"` |
| `BooleanParameterDefinition` | True/false toggle | `"value": "true"` or `"value": "false"` |
| `ChoiceParameterDefinition` | Predefined options | `"value": "Option"` (must match one of `choices`) |
| `PasswordParameterDefinition` | Hidden value | `"value": "secret"` |

#### Step 2: Build Parameters Array

Create the parameters JSON array with user-provided values:

```bash
# Example parameters
cat > params.json << 'EOF'
{
  "parameters": [
    {"name": "PERSON", "value": "stone"},
    {"name": "BIOGRAPHY", "value": "i'm a software engineer"},
    {"name": "TOGGLE", "value": "true"},
    {"name": "CHOICE", "value": "Two"},
    {"name": "PASSWORD", "value": "secret123"}
  ]
}
EOF
```

**Important:**
- Boolean values must be strings: `"true"` or `"false"`
- Use the `name` field, not the description
- Omit parameters to use their default values

#### Step 3: Trigger Build with Parameters

```bash
export BRANCH="main"

# Trigger with parameters
curl -s -X POST "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/pipelineruns?branch=${BRANCH}" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d @params.json | jq -r '.metadata.name'
```

#### Complete Parameterized Build Example

```bash
#!/bin/bash
export KUBESPHERE_API="http://kubesphere-apiserver:80"
export API_TOKEN="<tenant-oauth-token>"
export DEVOPS_PROJECT="devopstestc2nj7"
export PIPELINE_NAME="echo-jenkinsfile-pipeline"
export BRANCH="main"

# Step 1: Get parameters
echo "Fetching parameter definitions..."
PARAMS=$(curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches/${BRANCH}" \
  -H "Authorization: Bearer ${API_TOKEN}")

echo "Available parameters:"
echo "$PARAMS" | jq -r '.parameters[] | "- \(.name) (\(.type)): \(.description // "No description") [Default: \(.defaultParameterValue.value // "(none)")]"'

# Step 2: Ask user for values (or use defaults)
# In practice, prompt user or read from input

# Step 3: Build and trigger with parameters
echo "Triggering build with parameters..."
curl -s -X POST "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/pipelineruns?branch=${BRANCH}" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "parameters": [
      {"name": "PERSON", "value": "stone"},
      {"name": "BIOGRAPHY", "value": "software engineer"},
      {"name": "TOGGLE", "value": "true"},
      {"name": "CHOICE", "value": "Two"},
      {"name": "PASSWORD", "value": "secret123"}
    ]
  }' | jq -r '.metadata.name'
```

### Complete Working Example

```bash
#!/bin/bash
export KUBESPHERE_API="http://kubesphere-apiserver:80"
export API_TOKEN="<oauth-token>"
export DEVOPS_PROJECT="devopstestc2nj7"
export PIPELINE_NAME="echo-jenkinsfile-pipeline"

# Step 1: List branches
echo "Available branches:"
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches?filter=origin&page=1&limit=10" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" | jq -r '.items[] | "  - \(.name)"'

# Step 2: User selects branch (example: main)
BRANCH="main"

# Step 3: Trigger build
echo "Triggering build on branch: ${BRANCH}"
PIPELINE_RUN=$(curl -s -X POST "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/pipelineruns?branch=${BRANCH}" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"parameters":[]}')

RUN_NAME=$(echo "$PIPELINE_RUN" | jq -r '.metadata.name')
echo "PipelineRun created: ${RUN_NAME}"
```

---

## Alternative: API Endpoints (Deprecated v1alpha2)

> ⚠️ **Deprecated**: The `/kapis/devops.kubesphere.io/v1alpha2/` APIs are deprecated. Use `v1alpha3` APIs shown above.

Use APIs when you need programmatic access from external systems:

### Pipeline CRUD Operations

| Operation | Method | Endpoint | Status |
|-----------|--------|----------|--------|
| List Pipelines | GET | `/kapis/devops.kubesphere.io/v1alpha2/search?q=type:pipeline` | Deprecated |
| Get Pipeline | GET | `/kapis/devops.kubesphere.io/v1alpha2/namespaces/{devops}/pipelines/{pipeline}` |
| List Runs | GET | `/kapis/devops.kubesphere.io/v1alpha2/namespaces/{devops}/pipelines/{pipeline}/runs` |
| Run Pipeline (API) | POST | `/kapis/devops.kubesphere.io/v1alpha2/namespaces/{devops}/pipelines/{pipeline}/runs` |
| Get Run | GET | `/kapis/devops.kubesphere.io/v1alpha2/namespaces/{devops}/pipelines/{pipeline}/runs/{run}` |
| Stop Run | POST | `/kapis/devops.kubesphere.io/v1alpha2/namespaces/{devops}/pipelines/{pipeline}/runs/{run}/stop` |
| Get Log | GET | `/kapis/devops.kubesphere.io/v1alpha2/namespaces/{devops}/pipelines/{pipeline}/runs/{run}/log?start=0` | Use `?start=0` for tenant access |
| Get Artifacts | GET | `/kapis/devops.kubesphere.io/v1alpha2/namespaces/{devops}/pipelines/{pipeline}/runs/{run}/artifacts` |

### API Run Example

```bash
curl -X POST "https://kubesphere-api/kapis/devops.kubesphere.io/v1alpha2/namespaces/{devops}/pipelines/{pipeline}/runs" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "parameters": [
      {"name": "BRANCH", "value": "main"}
    ]
  }'
```

## Pipeline Run Flow

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  Queue   │───▶│  Running │───▶│ Complete │───▶│  Success │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
                      │                │
                      ▼                ▼
               ┌──────────┐     ┌──────────┐
               │ Paused   │     │  Failed  │
               │ (Input)  │     │          │
               └──────────┘     └──────────┘
```

## Pipeline Status Values

| Status | Meaning |
|--------|---------|
| **QUEUED** | Waiting for available agent |
| **RUNNING** | Currently executing |
| **PAUSED** | Waiting for user input |
| **SUCCESS** | Completed successfully |
| **FAILED** | Failed during execution |
| **ABORTED** | Manually stopped |

## Handling Paused Pipeline Steps with Input

When a pipeline reaches an `input` step (e.g., approval gates), it enters the **PAUSED** state and waits for user interaction. Tenants can approve or reject these steps via the KubeSphere API.

### Step 1: Get Node Details to Find Input Step

Use the `nodedetails` endpoint to identify the paused step with input:

```bash
# Get node details for the pipeline run
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelineruns/${PIPELINE_RUN_NAME}/nodedetails" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" | jq '.[-1]'
```

**Look for steps with `approvable: true`:**

```json
{
  "displayName": "Build Binary",
  "id": "38",
  "state": "PAUSED",
  "steps": [
    {
      "id": "41",
      "displayName": "Wait for interactive input",
      "state": "PAUSED",
      "input": {
        "id": "Build-binary-confirm",
        "message": "Build binary now?",
        "ok": "Proceed"
      },
      "approvable": true
    }
  ]
}
```

**Key Fields:**
| Field | Description |
|-------|-------------|
| `steps[].id` | Step ID (e.g., "41") |
| `steps[].input.id` | Input action ID (e.g., "Build-binary-confirm") |
| `steps[].approvable` | `true` if this step can be approved/rejected |

### Step 2: Approve or Reject the Input Step

Use the v1alpha2 API to submit your decision:

**Approve (Proceed):**
```bash
curl -s -X POST "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches/${BRANCH}/runs/${JENKINS_RUN_ID}/nodes/${NODE_ID}/steps/${STEP_ID}/" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"id":"${INPUT_ID}","abort":false}'
```

**Reject (Abort):**
```bash
curl -s -X POST "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches/${BRANCH}/runs/${JENKINS_RUN_ID}/nodes/${NODE_ID}/steps/${STEP_ID}/" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"id":"${INPUT_ID}","abort":true}'
```

**Parameters:**
| Parameter | Description | Example |
|-----------|-------------|---------|
| `JENKINS_RUN_ID` | Jenkins build number | `9` |
| `NODE_ID` | Stage/node ID from nodedetails | `38` |
| `STEP_ID` | Step ID from nodedetails | `41` |
| `INPUT_ID` | Input ID from the `input` object | `Build-binary-confirm` |
| `abort` | `false` to proceed, `true` to abort | `false` |

### Complete Example

```bash
export KUBESPHERE_API="http://kubesphere-apiserver:80"
export API_TOKEN="<tenant-oauth-token>"
export DEVOPS_PROJECT="devopstestc2nj7"
export PIPELINE_NAME="echo-jenkinsfile-pipeline"
export BRANCH="main"
export PIPELINE_RUN_NAME="echo-jenkinsfile-pipeline-ndrkn"

# Step 1: Get nodedetails and extract input information
NODES=$(curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelineruns/${PIPELINE_RUN_NAME}/nodedetails" \
  -H "Authorization: Bearer ${API_TOKEN}")

# Extract the last node (current stage) and find approvable step
NODE_ID=$(echo "$NODES" | jq -r '.[-1].id')
STEP_ID=$(echo "$NODES" | jq -r '.[-1].steps[] | select(.approvable == true) | .id')
INPUT_ID=$(echo "$NODES" | jq -r '.[-1].steps[] | select(.approvable == true) | .input.id')
JENKINS_RUN_ID=$(echo "$NODES" | jq -r '.[-1].steps[] | select(.approvable == true) | .jenkinsRunId // "9"')

echo "Approving step ${STEP_ID} in node ${NODE_ID}"
echo "Input ID: ${INPUT_ID}"

# Step 2: Approve the step
curl -s -X POST "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches/${BRANCH}/runs/${JENKINS_RUN_ID}/nodes/${NODE_ID}/steps/${STEP_ID}/" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"${INPUT_ID}\",\"abort\":false}"

# Step 3: Verify the pipeline continues
echo "Pipeline resumed. Check status..."
```

**Important Notes:**
- This approach works for **tenants** (no cluster-admin access required)
- Uses **v1alpha2** API for the approval action (not available in v1alpha3)
- The trailing slash `/` in the URL is required
- Empty response usually indicates success (HTTP 200)

## Retrieving Build Logs via KubeSphere API

To retrieve build logs through the KubeSphere API (recommended for tenants):

### Step 1: Get Jenkins PipelineRun ID

First, get the Jenkins PipelineRun ID from the PipelineRun CR annotation:

```bash
# Get fresh token
TOKEN_RESPONSE=$(curl -s -X POST "${KUBESPHERE_API}/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode 'grant_type=password' \
  --data-urlencode "username=${USERNAME}" \
  --data-urlencode "password=${PASSWORD}" \
  --data-urlencode 'client_id=kubesphere' \
  --data-urlencode 'client_secret=kubesphere')

export API_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')

# Get Jenkins PipelineRun ID from annotation
JENKINS_ID=$(curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelineruns/${PIPELINE_RUN_NAME}" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  | jq -r '.metadata.annotations."devops.kubesphere.io/jenkins-pipelinerun-id"')

echo "Jenkins Run ID: $JENKINS_ID"
```

### Step 2: Fetch Logs

Use the v1alpha2 API to fetch logs. **Note:** The `?start=0` parameter is required for tenant access:

```bash
# For multi-branch pipelines
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches/${BRANCH}/runs/${JENKINS_ID}/log/?start=0" \
  -H "Authorization: Bearer ${API_TOKEN}"

# Example:
curl -s "${KUBESPHERE_API}/clusters/member-1/kapis/devops.kubesphere.io/v1alpha2/namespaces/devopstestc2nj7/pipelines/echo-jenkinsfile-pipeline/branches/main/runs/5/log/?start=0" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

**Key Points:**
- Use the `jenkins-pipelinerun-id` annotation (not the Kubernetes PipelineRun name)
- For multi-branch pipelines: include `/branches/{branch}/` in the path
- **CRITICAL:** Include `?start=0` query parameter - required for tenant access
- The trailing slash `/` at the end of the URL path is required
- Works with tenant credentials (no cluster-admin access needed)

## Retrieving Build Artifacts via KubeSphere API

To query and download build artifacts through the KubeSphere API:

### Step 1: List Artifacts

Query artifacts for a specific run:

```bash
curl -s "${KUBESPHERE_API}/clusters/${CLUSTER}/kapis/devops.kubesphere.io/v1alpha2/namespaces/${DEVOPS_PROJECT}/pipelines/${PIPELINE_NAME}/branches/${BRANCH}/runs/${JENKINS_ID}/artifacts/?start=0&limit=10" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

**Example Response:**
```json
[
  {
    "name": "service",
    "path": "service",
    "size": 8344228,
    "downloadable": true,
    "url": "/job/devopstestc2nj7/job/go-pipeline/job/main/2/artifact/service"
  }
]
```

### Step 2: Download Artifact

Download a specific artifact by filename:

```bash
curl -s -o service "${KUBESPHERE_API}/kapis/clusters/${CLUSTER}/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelineruns/${PIPELINE_RUN_NAME}/artifacts/download?filename=service" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

**Key Points:**
- Use the **Kubernetes PipelineRun name** (not Jenkins ID) for the download endpoint
- The filename must match exactly what's listed in the artifacts query
- Supports pagination with `start` and `limit` parameters for listing
- Works with tenant credentials (no cluster-admin access needed)

## Querying Jenkins Directly

For debugging, you can query Jenkins directly using the admin token:

```bash
# Get Jenkins admin token
TOKEN=$(kubectl -n kubesphere-devops-system get secret devops-jenkins -o jsonpath='{.data.jenkins-admin-token}' | base64 -d)

# List all folders (maps to DevOpsProjects)
kubectl run curl-jenkins --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/api/json" \
  -H "Content-Type: application/json"

# List jobs in a specific folder
kubectl run curl-jenkins --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/api/json" \
  -H "Content-Type: application/json"
```

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| **PipelineRun not triggering** | Check Pipeline exists and is synced to Jenkins |
| **Controller panic (getAgentInfo nil pointer)** | Known issue with PipelineRef; use Jenkins API directly as workaround |
| **Agent label not found** | Check available labels: `kubectl run curl-jenkins --rm -i --restart=Never --image=curlimages/curl -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/labelsdashboard/labelsData"` |
| **Go/Tool not found in agent** | Use `kubernetes { yaml ... }` agent with appropriate container image |
| **Artifact not found** | Run `archiveArtifacts` in the same stage where the file was created (same workspace) |
| **Permission denied** | Check DevOps project membership and RBAC |
| **Pipeline shows in KubeSphere but not Jenkins** | Check sync status annotation: `pipeline.devops.kubesphere.io/syncstatus` |
| **Run fails immediately** | Check controller logs: `kubectl logs -n kubesphere-devops-system deployment/devops-controller` |
| **Escape characters in Jenkinsfile** | Don't use `\$class` escape sequences in inline Jenkinsfile |

### Debugging Steps

1. **Check PipelineRun status:**
   ```bash
   kubectl get pipelinerun <name> -n <namespace> -o yaml
   ```

2. **Check Jenkins build directly:**
   ```bash
   # Get admin token
   TOKEN=$(kubectl -n kubesphere-devops-system get secret devops-jenkins -o jsonpath='{.data.jenkins-admin-token}' | base64 -d)
   
   # Get console output
   kubectl run curl-jenkins --rm -i --restart=Never --image=curlimages/curl \
     -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/<project>/job/<pipeline>/<build>/consoleText"
   ```

3. **List Jenkins agent labels:**
   ```bash
   kubectl run curl-jenkins --rm -i --restart=Never --image=curlimages/curl \
     -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/labelsdashboard/labelsData"
   ```

| Mistake | Fix |
|---------|-----|
| PipelineRun not triggering | Check Pipeline exists and is synced to Jenkins |
| Agent not found | Verify Jenkins agent labels match pipeline's `agent { label 'xxx' }` |
| Permission denied | Check DevOps project membership and RBAC |
| Pipeline shows in KubeSphere but not Jenkins | Check sync status annotation: `pipeline.devops.kubesphere.io/syncstatus` |
| Run fails immediately | Check controller logs: `kubectl logs -n kubesphere-devops-system deployment/devops-controller` |

## Best Practices

1. **Use PipelineRun CRDs** for triggering - it's the cloud-native approach
2. **Check status via kubectl** rather than polling APIs
3. **Use `kubectl logs`** on the PipelineRun controller for debugging
4. **Verify sync status** - pipelines must sync to Jenkins before they can run

## References

- [DevOps API Swagger](/root/go/src/github.com/kubesphere/kse-extensions/devops/swagger.yaml)
- [Jenkins Pipeline Syntax](https://www.jenkins.io/doc/book/pipeline/syntax/)
