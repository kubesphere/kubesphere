---
name: kubesphere-fluid
description: KubeSphere Fluid management Skill. Use when user asks to install or enable Fluid, check Fluid status, view Fluid pods/logs/CRDs, create or update Dataset, AlluxioRuntime, JuiceFSRuntime, or ThinRuntime, perform DataLoad or cache warming, scale runtime, or troubleshoot Fluid issues in KubeSphere.
---

# KubeSphere Fluid Management

Use this skill for the full Fluid lifecycle in KubeSphere:

- Install or upgrade the Fluid extension through `InstallPlan`
- Query extension status, CRDs, Pods, and logs
- Generate `Dataset` manifests with mount configuration
- Generate `AlluxioRuntime`, `JuiceFSRuntime`, or `ThinRuntime` manifests for caching
- Perform DataLoad for cache warming
- Scale runtime replicas
- Uninstall the Fluid extension
- Troubleshoot Fluid Pods, CRDs, and mount issues

Out of scope by default:

- Advanced Fluid operations not requested by the user, such as `DataBackup` or `GooseFS`
- Deep tiered storage configuration without user requirements

If the user explicitly asks for those, acknowledge that they are Fluid capabilities but treat them as a follow-up task.

## Response Rules

- Prefer executable output: `InstallPlan` YAML, `Dataset` YAML, `AlluxioRuntime` YAML, kubectl commands, or a short ordered procedure.
- Verify the extension name and exact version before generating a final `InstallPlan`.
- Default the extension name to `fluid` only if the user context or cluster output does not expose a different resource name.
- Use `upgradeStrategy: Manual` unless the user explicitly asks for something else.
- Omit optional fields instead of guessing values.
- Before generating Dataset or Runtime, prefer checking the installed API versions:

```bash
kubectl api-resources --api-group data.fluid.io
```

- If the cluster version is unknown and the user only wants an example, use:
  - Dataset: `data.fluid.io/v1alpha1`
  - AlluxioRuntime: `data.fluid.io/v1alpha1`
  - JuiceFSRuntime: `data.fluid.io/v1alpha1`
  - ThinRuntime: `data.fluid.io/v1alpha1`
- For uninstall requests, warn that deleting CRDs or CR instances can remove application configuration.

## CRITICAL: Parameter Handling

**ALWAYS use the exact values provided by the user. Never substitute or guess values.**

When generating YAML manifests:

| Parameter | Rule |
|-----------|------|
| `name` | MUST use user's value exactly |
| `namespace` | MUST use user's value exactly |
| `mountPoint` | MUST use user's value exactly (e.g., s3://bucket, oss://bucket, pvc://, https://) |
| `replicas` | MUST use user's value exactly |
| `quota` | MUST use user's value exactly |
| `mediumType` | Use user's value, default to MEM if not specified |
| `path` | Use user's value, default to /dev/shm if not specified |

**WRONG**: Using `pvc://` when user specified `s3://mybucket/spark-data`
**RIGHT**: Using exactly what user provided: `s3://mybucket/spark-data`

## CRITICAL: Operation Scope

**Each operation has a specific scope. Do NOT create additional resources unless user explicitly asks.**

| User Request | Output Scope |
|--------------|--------------|
| "Create Dataset" | Only Dataset YAML |
| "Create AlluxioRuntime" | Only AlluxioRuntime YAML (NOT Dataset) |
| "Create JuiceFSRuntime" | Only JuiceFSRuntime YAML (NOT Dataset) |
| "Create ThinRuntime" | Only ThinRuntime YAML (NOT Dataset) |
| "Create Dataset with Runtime" | Both Dataset + Runtime YAML |
| "Create DataLoad" | Only DataLoad YAML |

**Example:**
- User says: "Create an AlluxioRuntime" → Output ONLY AlluxioRuntime YAML
- User says: "Create a Dataset with AlluxioRuntime" → Output both Dataset and AlluxioRuntime YAML

## Version Mapping Discovery

When the user asks for precision, prove the version mapping first:

```bash
# Discover KubeSphere extension version
kubectl get extensionversions.kubesphere.io -l kubesphere.io/extension-ref=fluid
kubectl get extensionversion fluid-<version> -o yaml

# Discover Fluid runtime image or controller version
kubectl get pods -n fluid-system -o wide
kubectl get deploy -n fluid-system alluxio-runtime-controller -o jsonpath='{.spec.template.spec.containers[*].image}'
kubectl describe pod -n fluid-system <fluid-pod>
```

If these commands disagree with assumed mappings, prefer cluster output over defaults.

## Discovery Commands

This section provides three approaches for querying Fluid status:
1. **KubeSphere API (curl)** - for extension management and multi-cluster queries
2. **kubectl** - for direct Kubernetes resource operations
3. **ksctl CLI** - KubeSphere command-line tool

### Option 1: Using KubeSphere API (curl)

Use curl with environment variables for querying KubeSphere extension status and multi-cluster resources.

**Environment Variables:**
```bash
export KS_HOST="http://<kubesphere-host>"     # KubeSphere console URL (required)
export KS_USERNAME="admin"                     # Username (default: admin)
export KS_PASSWORD="<password>"                # Password (required)
```

**Helper Functions (add to ~/.bashrc or use directly):**
```bash
# Get OAuth token
ks_token() {
  curl -s -X POST "$KS_HOST/oauth/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=password&username=${KS_USERNAME:-admin}&password=$KS_PASSWORD&client_id=kubesphere&client_secret=kubesphere" | jq -r '.access_token'
}

# Make API call: ks_api GET/POST/PUT/DELETE <path> [body]
ks_api() {
  local method=${1:-GET}
  local path=$2
  local body=$3
  local token=$(ks_token)
  
  curl -s -X "$method" \
    -H "Authorization: Bearer $token" \
    -H "Content-Type: application/json" \
    ${body:+-d "$body"} \
    "$KS_HOST$path"
}
```

**Query Commands:**

```bash
# List all clusters (host + member clusters)
ks_api GET /kapis/cluster.kubesphere.io/v1alpha1/clusters | jq -r '.items[].metadata.name'

# List installed extensions
ks_api GET /kapis/kubesphere.io/v1alpha1/extensions | jq -r '.items[].metadata.name' | grep -i fluid

# List available extension versions
ks_api GET /kapis/kubesphere.io/v1alpha1/extensionversions | jq -r '.items[].metadata.name' | grep -i fluid

# Get Fluid extension details
ks_api GET /kapis/kubesphere.io/v1alpha1/extensions/fluid | jq

# Get cluster connection status
ks_api GET /kapis/cluster.kubesphere.io/v1alpha1/clusters/<cluster>/status | jq '.conditions'
```

**Multi-Cluster Resource Query:**

```bash
# Get datasets in host cluster
ks_api GET /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/<namespace>/datasets

# Get alluxioruntimes in host cluster
ks_api GET /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/<namespace>/alluxioruntimes

# Get all runtimes (Alluxio + JuiceFS + Thin)
ks_api GET /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/<namespace>/allruntimes

# Get specific dataset
ks_api GET /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/<namespace>/datasets/<name>

# Get CRDs in cluster
ks_api GET /clusters/host/apis/apiextensions.k8s.io/v1/customresourcedefinitions | jq -r '.items[] | select(.metadata.name | contains("fluid")) | .metadata.name'
```

**API Path Format:**
```
# For KubeSphere extension management:
/kapis/kubesphere.io/v1alpha1/extensions
/kapis/cluster.kubesphere.io/v1alpha1/clusters

# For Fluid resources in namespace:
/clusters/{cluster}/kapis/data.fluid.io/v1alpha1/namespaces/{namespace}/{resources}
/clusters/{cluster}/kapis/data.fluid.io/v1alpha1/namespaces/{namespace}/{resources}/{name}
```

**Query Parameters:**
- `page` - Page number (default: 1)
- `limit` - Items per page
- `ascending` - Sort direction (default: false)
- `sortBy` - Sort field (e.g., createTime)

**Supported Resource Types:**
- `datasets`
- `alluxioruntimes`
- `juicefsruntimes`
- `dataloads`
- `thinruntimes`
- `allruntimes`

### CRUD Operations via KubeSphere API

**Create Operations:**

```bash
# Create Dataset
DATASET_JSON='{
  "apiVersion": "data.fluid.io/v1alpha1",
  "kind": "Dataset",
  "metadata": {
    "name": "<name>",
    "namespace": "<namespace>"
  },
  "spec": {
    "mounts": [{
      "mountPoint": "<mountPoint>",
      "name": "<mountName>"
    }]
  }
}'

ks_api POST /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/$NAMESPACE/datasets "$DATASET_JSON"

# Or use kubectl:
# kubectl apply -f examples/dataset.yaml
```

**Read Operations:**

```bash
# List all datasets in namespace
ks_api GET /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/$NAMESPACE/datasets

# Get specific dataset
ks_api GET /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/$NAMESPACE/datasets/<name>
```

**Update Operations (Scale Runtime):**

```bash
# Scale AlluxioRuntime via PUT
RUNTIME_UPDATE='{"spec":{"replicas":<new-replicas>}}'
ks_api PUT /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/$NAMESPACE/alluxioruntimes/<name> "$RUNTIME_UPDATE"

# Or use kubectl:
# kubectl scale alluxioruntime <name> -n <namespace> --replicas=<n>
```

**Delete Operations:**

```bash
# Delete dataset
ks_api DELETE /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/$NAMESPACE/datasets/<name>

# Delete alluxioruntime
ks_api DELETE /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/$NAMESPACE/thinruntimes/<name>
ks_api DELETE /clusters/host/kapis/data.fluid.io/v1alpha1/namespaces/$NAMESPACE/alluxioruntimes/<name>
```

### Option 2: Using kubectl (Direct Cluster Access)

Use `kubectl` for direct Kubernetes resource operations.

```bash
# Extension and version discovery in KubeSphere
kubectl get extensions.kubesphere.io | grep -i fluid
kubectl get extensionversions.kubesphere.io | grep -i fluid

# InstallPlan and extension status
kubectl get installplans.kubesphere.io
kubectl get installplan fluid -o yaml
kubectl describe extension fluid
kubectl describe extensionversion fluid-<version>

# Fluid runtime resources
kubectl api-resources --api-group data.fluid.io
kubectl get crd | grep -E 'datasets.data.fluid.io|alluxioruntimes.data.fluid.io'
kubectl get pods -A | grep -i fluid
kubectl get datasets.data.fluid.io -A
kubectl get alluxioruntimes.data.fluid.io -A
```

### Option 3: Using ksctl CLI

```bash
# List available datasets
ksctl get dataset -n <namespace>

# Create dataset with runtime
ksctl create dataset -f dataset.yaml

# Check status
ksctl describe dataset <name> -n <namespace>
```

### When to Use Which Approach

| Scenario | Recommended Approach |
|----------|---------------------|
| Query KubeSphere extension status | KubeSphere API (curl) |
| List available clusters | KubeSphere API (curl) |
| Query host cluster Kubernetes resources | kubectl |
| Query member cluster Kubernetes resources | kubeconfig extraction |
| Create/apply Dataset/Runtime | kubectl |
| Get extension version info | KubeSphere API (curl) |
| Quick status check | ksctl |

KubeSphere UI relation:

- Users can also discover and enable the extension from the KubeSphere Extension Marketplace UI
- When the user asks for repeatable or reviewable operations, prefer `InstallPlan` manifests and `kubectl`
- When the user explicitly wants console steps, explain the equivalent UI path instead of forcing YAML

---

## Example Files

This skill includes ready-to-use example manifests in the `examples/` directory:

| File | Description |
|------|-------------|
| `examples/dataset.yaml` | Minimal Dataset manifest |
| `examples/alluxioruntime.yaml` | Dataset + AlluxioRuntime with tiered storage |
| `examples/juicefs.yaml` | Dataset + JuiceFSRuntime |
| `examples/thinruntime.yaml` | Dataset + ThinRuntime |
| `examples/dataload.yaml` | DataLoad for cache warming |
| `examples/installplan.yaml` | InstallPlan for Fluid extension |

---

## Install Fluid in KubeSphere

### 1. Pre-check

```bash
kubectl get extension fluid
kubectl get extensionversions.kubesphere.io -l kubesphere.io/extension-ref=fluid
kubectl describe extensionversion fluid-<exact-version>
```

Check:

- The extension resource exists
- The exact extension version exists
- Whether the extension is single-cluster or multi-cluster
- Whether the user needs custom config

### 2. InstallPlan Template

Use this when the user has already confirmed the extension name and version.

```bash
# See examples/installplan.yaml for the template
kubectl apply -f examples/installplan.yaml
kubectl get installplan fluid -w
kubectl describe installplan fluid
kubectl get extension fluid -o yaml
```

---

## Manage Dataset

### Dataset Template

```bash
# See examples/dataset.yaml
# Key fields:
# - spec.mounts[].mountPoint: Data source (s3://, oss://, pvc://, https://) - USE EXACT VALUE FROM USER
# - spec.mounts[].name: Mount identifier
# - spec.mounts[].readOnly: Read-only mount (default: false)
```

**IMPORTANT**: When user provides `mountPoint`, use it EXACTLY. Do not substitute.

Common operations:

```bash
kubectl apply -f examples/dataset.yaml
kubectl get dataset <name> -n <namespace>
kubectl describe dataset <name> -n <namespace>
kubectl delete dataset <name> -n <namespace>
```

---

## Manage AlluxioRuntime

### AlluxioRuntime Template (Standalone)

**When user asks to create AlluxioRuntime ONLY, generate ONLY AlluxioRuntime YAML:**

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: {{name}}
  namespace: {{namespace}}
spec:
  replicas: {{replicas}}
  tieredstore:
    levels:
      - mediumtype: {{mediumType}}
        path: {{path}}
        quota: {{quota}}
        high: "{{high}}"
        low: "{{low}}"
```

**When user explicitly asks for "Dataset with Runtime", generate both:**

```bash
# See examples/alluxioruntime.yaml
```

### Key fields:
- `spec.replicas`: Number of Alluxio workers - USE USER'S VALUE
- `spec.tieredstore.levels`: Storage configuration
  - `mediumtype`: MEM, SSD, or HDD - USE USER'S VALUE or default to MEM
  - `path`: Storage path - USE USER'S VALUE or default to /dev/shm
  - `quota`: Storage size - USE USER'S VALUE
  - `high/low`: Watermark ratios

Common operations:

```bash
kubectl apply -f examples/alluxioruntime.yaml
kubectl get alluxioruntime <name> -n <namespace> -o wide
kubectl describe alluxioruntime <name> -n <namespace>
```

### Scale Runtime

```bash
# Scale via kubectl
kubectl scale alluxioruntime <name> -n <namespace> --replicas=<n>

# Scale via kubectl patch
kubectl patch alluxioruntime <name> -n <namespace> -p '{"spec":{"replicas":<n>}}'

# Verify
kubectl get alluxioruntime <name> -n <namespace> -o wide
```

---

## Manage JuiceFSRuntime

### JuiceFSRuntime Template (Standalone)

**When user asks to create JuiceFSRuntime ONLY, generate ONLY JuiceFSRuntime YAML:**

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: JuiceFSRuntime
metadata:
  name: {{name}}
  namespace: {{namespace}}
spec:
  volume:
    name: {{juicefsVolume}}
    secret: {{secretName}}
```

**When user explicitly asks for "Dataset with JuiceFS", generate both:**

```bash
# See examples/juicefs.yaml
```

### Key fields:
- `spec.volume.name`: JuiceFS volume name - USE USER'S VALUE
- `spec.volume.secret`: Credentials secret - USE USER'S VALUE

Common operations:

```bash
kubectl apply -f examples/juicefs.yaml
kubectl get juicefsruntime <name> -n <namespace> -o wide
kubectl describe juicefsruntime <name> -n <namespace>
```

---

## DataLoad (Cache Warming)

### DataLoad Template - COMPLETE VERSION

**ALWAYS include all required fields:**

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: {{name}}
  namespace: {{namespace}}
spec:
  dataset:
    name: {{datasetName}}        # REQUIRED - target dataset name
    namespace: {{datasetNamespace}}  # REQUIRED - target dataset namespace
  loadMetadata: {{loadMetadata}} # REQUIRED - boolean (default: false)
  target:                        # REQUIRED - array of paths to load
    - path: {{targetPath}}       # Path to load, e.g., /data
```

**Field requirements:**
- `spec.dataset.name`: REQUIRED - must be provided by user
- `spec.dataset.namespace`: REQUIRED - must be provided by user  
- `spec.loadMetadata`: REQUIRED - default to `false` if user doesn't specify
- `spec.target`: REQUIRED - array of paths, minimum one entry
- `spec.target[].path`: REQUIRED - the path to load (e.g., /data, /user/home)

Common operations:

```bash
kubectl apply -f examples/dataload.yaml
kubectl get dataload <name> -n <namespace>
kubectl describe dataload <name> -n <namespace>
```

---

## Uninstall Fluid

### Uninstall Pre-check

Check whether datasets or runtimes still exist:

```bash
kubectl get datasets.data.fluid.io -A
kubectl get alluxioruntimes.data.fluid.io -A
kubectl get juicefsruntimes.data.fluid.io -A
kubectl get thinruntimes.data.fluid.io -A
```

If these resources still exist, tell the user to migrate or remove them first.

### Default uninstall path:

```bash
kubectl delete installplan fluid
kubectl get installplan fluid
kubectl get pods -n fluid-system | grep -i fluid
```

If the user wants full cleanup, remind them to handle application resources first.

---

## Troubleshooting Playbook

### 1. InstallPlan or extension failed

```bash
kubectl describe installplan fluid
kubectl get installplan fluid -o jsonpath='{.status.conditions}'
kubectl get extension fluid -o yaml
kubectl get extensionversion fluid-<version> -o yaml
```

### 2. Fluid Pods are unhealthy

```bash
kubectl get pods -A | grep -i fluid
kubectl describe pod -n fluid-system <pod-name>
kubectl logs -n fluid-system deploy/alluxio-runtime-controller --tail=200
kubectl logs -n fluid-system deploy/juicefs-runtime-controller --tail=200
kubectl get events -n fluid-system --sort-by=.lastTimestamp
```

### 3. CRD missing or not established

```bash
kubectl get crd datasets.data.fluid.io alluxioruntimes.data.fluid.io juicefsruntimes.data.fluid.io thinruntimes.data.fluid.io
kubectl describe crd datasets.data.fluid.io
kubectl api-resources --api-group data.fluid.io
```

Likely causes:

- Fluid extension installation did not finish
- CRDs were partially created or rejected by the API server
- Controller startup failed before CRDs became usable

Safe next actions:

- Re-check `InstallPlan` state and conditions
- Inspect install logs and controller logs

### 4. Runtime Not Ready

```bash
kubectl get alluxioruntime <name> -n <namespace>
kubectl describe alluxioruntime <name> -n <namespace>
kubectl get pods -n <namespace> -l release=<runtime>
```

Likely causes:

- Master or worker pods not scheduled
- Tiered storage configuration issue
- Insufficient node resources

### 5. Mount Failed

```bash
kubectl describe dataset <name> -n <namespace>
kubectl get dataset <name> -n <namespace> -o yaml
```

Likely causes:

- Mount point inaccessible
- Invalid credentials
- Network connectivity to UFS

### 6. DataLoad Stuck

```bash
kubectl get dataload <name> -n <namespace>
kubectl describe dataload <name> -n <namespace>
kubectl get dataset <dataset-name> -n <namespace>
```

### 7. Scaling Slow

```bash
kubectl describe alluxioruntime <name> -n <namespace> | grep -A5 "Scaling"
kubectl describe resourcequota -n <namespace>
kubectl describe nodes | grep -A10 "Allocated resources"
```

---

## Output Patterns

Match the answer to the user intent:

- Install request: brief pre-check commands plus a complete `InstallPlan` manifest
- Status request: only the most relevant `kubectl` commands, grouped by purpose
- Dataset request: one executable manifest plus apply, verify, and delete commands
- AlluxioRuntime request: ONLY AlluxioRuntime manifest (NOT Dataset), unless user explicitly says "with Dataset"
- Runtime with Dataset request: both Dataset and Runtime manifests
- DataLoad request: COMPLETE manifest with all required fields (dataset, loadMetadata, target)
- Scale request: scale command and verification
- Uninstall request: delete command, verification commands, and cleanup cautions
- Troubleshooting request: diagnosis commands first, then likely causes, then safe next steps

## Manage ThinRuntime

### ThinRuntime Template (Standalone)

**When user asks to create ThinRuntime ONLY, generate ONLY ThinRuntime YAML:**

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: {{name}}
  namespace: {{namespace}}
spec:
  mountPoint: {{mountPoint}}
 thin:
    profile: {{profileName}}
    credentials: {{secretName}}
```

**When user explicitly asks for "Dataset with ThinRuntime", generate both:**

```bash
# See examples/thinruntime.yaml
```

### Key fields:
- `spec.mountPoint`: Under storage path - USE USER'S VALUE
- `spec.thin.profile`: Thin runtime profile name - USE USER'S VALUE
- `spec.thin.credentials`: Secret containing credentials - USE USER'S VALUE

Common operations:

```bash
kubectl apply -f examples/thinruntime.yaml
kubectl get thinruntime <name> -n <namespace> -o wide
kubectl describe thinruntime <name> -n <namespace>
```

---

