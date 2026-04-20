---
name: kubesphere-openkruise
description: KubeSphere OpenKruise management Skill. Use when user asks to install or enable OpenKruise, check OpenKruise status, view kruise pods/logs/CRDs, create or update SidecarSet, manage sidecar injection, create or update CloneSet, perform in-place update or batch rollout, uninstall or remove OpenKruise, or troubleshoot Kruise Pod, CRD, and webhook issues in KubeSphere.
---

# Skill: kubesphere-openkruise

# KubeSphere OpenKruise Management

Use this skill for the full OpenKruise lifecycle in KubeSphere:

- Install or upgrade the OpenKruise extension through `InstallPlan`
- Query extension status, CRDs, Pods, and logs
- Generate `SidecarSet` manifests for sidecar injection and rolling updates
- Generate `CloneSet` manifests for advanced stateless workloads, in-place updates, and batch rollout
- Uninstall the OpenKruise extension
- Troubleshoot failed Pods, missing CRDs, and webhook problems

Out of scope by default:

- Advanced OpenKruise CRDs not requested by the user, such as `BroadcastJob`, `NodeImage`, `UnitedDeployment`, or `AdvancedPodAutoscaler`
- Deep chart-value design without cluster evidence

If the user explicitly asks for those CRDs, acknowledge that they are OpenKruise capabilities but treat them as a follow-up task instead of assuming they belong in the default workflow.

## Response Rules

- Prefer executable output: `InstallPlan` YAML, `SidecarSet` YAML, `CloneSet` YAML, `kubectl` commands, or a short ordered procedure.
- Verify the extension name and exact version before generating a final `InstallPlan`.
- Default the extension name to `openkruise` only if the user context or cluster output does not expose a different resource name.
- In this KubeSphere environment, the observed mapping is:
  - KubeSphere extension version `1.0.3`
  - OpenKruise runtime version `1.4.0`
- Distinguish two version concepts before generating `InstallPlan`:
  - KubeSphere extension version: used by `spec.extension.version`
  - OpenKruise runtime version: the controller or component version seen in Pods or docs
- For this environment, if the user asks to install the current OpenKruise plugin and does not provide another extension version, prefer `spec.extension.version: 1.0.3`.
- If the user only provides the OpenKruise runtime version, first ask them to confirm the matching KubeSphere extension version.
- Never invent a version. If the version is missing, first show how to list versions and ask the user to confirm one.
- `InstallPlan.metadata.name` MUST equal `InstallPlan.spec.extension.name`.
- `InstallPlan` is cluster-scoped in KubeSphere. Do not add a namespace to `kubectl get|describe|delete installplan`.
- Use `upgradeStrategy: Manual` unless the user explicitly asks for something else.
- Omit optional fields instead of guessing values.
- Before generating `SidecarSet` or `CloneSet`, prefer checking the installed API versions:

```bash
kubectl api-resources --api-group apps.kruise.io
```

- If the cluster version is unknown and the user only wants an example, prefer:
  - `SidecarSet`: `apps.kruise.io/v1alpha1`
  - `CloneSet`: `apps.kruise.io/v1alpha1`
- For uninstall requests, warn that deleting CRDs or CR instances can remove application configuration. Do not suggest deleting CRDs unless the user explicitly asks for full cleanup.

## Version Mapping Discovery

Treat the `1.0.3 -> 1.4.0` mapping as environment evidence, not a universal rule. When the user asks for precision, prove it first:

```bash
# Discover KubeSphere extension version
kubectl get extensionversions.kubesphere.io -l kubesphere.io/extension-ref=openkruise
kubectl get extensionversion openkruise-1.0.3 -o yaml

# Discover deployed runtime image or controller version
kubectl get pods -n kruise-system -o wide
kubectl get deploy -n kruise-system kruise-manager -o jsonpath='{.spec.template.spec.containers[*].image}'
kubectl describe pod -n kruise-system <kruise-manager-pod>
```

If these commands disagree with the assumed mapping, prefer cluster output over the baked-in default.

## Discovery Commands

This section provides two approaches for querying OpenKruise status:
1. **KubeSphere API (curl)** - for extension management and multi-cluster queries
2. **kubectl** - for direct Kubernetes resource operations

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
ks_api GET /kapis/kubesphere.io/v1alpha1/extensions | jq -r '.items[].metadata.name' | grep -i kruise

# List available extension versions
ks_api GET /kapis/kubesphere.io/v1alpha1/extensionversions | jq -r '.items[].metadata.name' | grep -i kruise

# Get OpenKruise extension details
ks_api GET /kapis/kubesphere.io/v1alpha1/extensions/openkruise | jq

# Get cluster connection status (for member clusters)
ks_api GET /kapis/cluster.kubesphere.io/v1alpha1/clusters/member-4/status | jq '.conditions'
```

### Multi-Cluster Resource Query (host cluster)

```bash
# Get pods in host cluster
ks_api GET /clusters/host/api/v1/namespaces/kruise-system/pods | jq -r '.items[].metadata.name'

# Get SidecarSets in host cluster (cluster-wide)
ks_api GET /clusters/host/kapis/apps.kruise.io/v1alpha1/users/admin/sidecarsets | jq -r '.items[].metadata.name'

# Get CloneSets in host cluster (cluster-wide)
ks_api GET /clusters/host/kapis/apps.kruise.io/v1alpha1/users/admin/clonesets | jq -r '.items[].metadata.name'

# Get SidecarSets in specific namespace
ks_api GET /clusters/host/kapis/apps.kruise.io/v1alpha1/namespaces/default/sidecarsets | jq -r '.items[].metadata.name'

# Get namespaces where a SidecarSet is applied
ks_api GET /clusters/host/kapis/apps.kruise.io/v1alpha1/users/admin/sidecarsetname/side1 | jq -r '.items[].metadata.name'

# Get specific SidecarSet details
ks_api GET /clusters/host/kapis/apps.kruise.io/v1alpha1/users/admin/sidecarsets/side1 | jq

# Get CRDs in host cluster
ks_api GET /clusters/host/apis/apiextensions.k8s.io/v1/customresourcedefinitions | jq -r '.items[] | select(.metadata.name | contains("kruise")) | .metadata.name'
```

### Multi-Cluster Resource Query (member cluster)

```bash
# Get pods in member cluster
ks_api GET /clusters/member-4/api/v1/namespaces/kruise-system/pods | jq -r '.items[].metadata.name'

# Get SidecarSets in member cluster (cluster-wide)
ks_api GET /clusters/member-4/kapis/apps.kruise.io/v1alpha1/users/admin/sidecarsets | jq -r '.items[].metadata.name'

# Get CloneSets in member cluster (cluster-wide)
ks_api GET /clusters/member-4/kapis/apps.kruise.io/v1alpha1/users/admin/clonesets | jq -r '.items[].metadata.name'

# Get CRDs in member cluster
ks_api GET /clusters/member-4/apis/apiextensions.k8s.io/v1/customresourcedefinitions | jq -r '.items[] | select(.metadata.name | contains("kruise")) | .metadata.name'
```

**API Path Format:**

```
# For KubeSphere extension management:
/kapis/kubesphere.io/v1alpha1/extensions
/kapis/cluster.kubesphere.io/v1alpha1/clusters

# For cluster-wide OpenKruise resources (SidecarSet, CloneSet):
/clusters/{cluster}/kapis/apps.kruise.io/v1alpha1/users/admin/{resources}
/clusters/{cluster}/kapis/apps.kruise.io/v1alpha1/users/admin/{resources}/{name}
/clusters/{cluster}/kapis/apps.kruise.io/v1alpha1/users/admin/sidecarsetname/{sidecarsetName}

/clusters/{cluster}/api/v1/namespaces/{namespace}/{resources}
/clusters/{cluster}/apis/apiextensions.k8s.io/v1/customresourcedefinitions
```

**Query Parameters:**
- `page` - Page number (default: 1)
- `limit` - Items per page
- `ascending` - Sort direction (default: false)
- `sortBy` - Sort field (e.g., createTime)

### Option 2: Using kubectl (Direct Cluster Access)

Use `kubectl` for direct Kubernetes resource operations on the **host cluster**.

```bash
# Extension and version discovery in KubeSphere
kubectl get extensions.kubesphere.io | grep -i kruise
kubectl get extensionversions.kubesphere.io | grep -i kruise
kubectl get extensionversions.kubesphere.io -l kubesphere.io/extension-ref=openkruise

# InstallPlan and extension status
kubectl get installplans.kubesphere.io
kubectl get installplan openkruise -o yaml
kubectl describe extension openkruise
kubectl describe extensionversion openkruise-<version>

# OpenKruise runtime resources
kubectl api-resources --api-group apps.kruise.io
kubectl get crd | grep -E 'clonesets.apps.kruise.io|sidecarsets.apps.kruise.io'
kubectl get pods -A | grep -i kruise
kubectl get sidecarsets.apps.kruise.io -A
kubectl get clonesets.apps.kruise.io -A
```

### Option 3: Multi-Cluster Query (kubeconfig extraction)

For querying **member clusters**, extract the kubeconfig from the Cluster resource:

```bash
# Get kubeconfig for a member cluster
CLUSTER_NAME=member-4
KUBECONFIG_ENCODED=$(kubectl get cluster.cluster.kubesphere.io $CLUSTER_NAME -o jsonpath='{.spec.connection.kubeconfig}')
echo "$KUBECONFIG_ENCODED" | base64 -d > /tmp/${CLUSTER_NAME}-kubeconfig

# Query member cluster
export KUBECONFIG=/tmp/${CLUSTER_NAME}-kubeconfig
kubectl get pods -n kruise-system
kubectl get sidecarsets.apps.kruise.io -A
kubectl get clonesets.apps.kruise.io -A

# Switch back to host cluster
export KUBECONFIG=""
```

### When to Use Which Approach

| Scenario | Recommended Approach |
|----------|---------------------|
| Query KubeSphere extension status | KubeSphere API (curl) |
| List available clusters | KubeSphere API (curl) |
| Query host cluster Kubernetes resources | kubectl |
| Query member cluster Kubernetes resources | kubeconfig extraction |
| Create/apply InstallPlan/SidecarSet/CloneSet | kubectl |
| Get extension version info | KubeSphere API (curl) |

## Install OpenKruise in KubeSphere

### 1. Pre-check

```bash
kubectl get extension openkruise
kubectl get extensionversions.kubesphere.io -l kubesphere.io/extension-ref=openkruise
kubectl describe extensionversion openkruise-<exact-version>
```

Check:

- The extension resource really exists
- The exact extension version exists
- Whether the extension is single-cluster or multi-cluster
- Whether the user needs custom config
- Whether the user is giving an extension version or only the OpenKruise runtime version
- In this environment, remember that extension `1.0.3` maps to OpenKruise runtime `1.4.0`

### 2. InstallPlan Template

Use this when the user has already confirmed the extension name and version.

**Note:** 
- InstallPlan is a KubeSphere CRD resource, **created and managed only in the host cluster** (not in member clusters)
- `clusterScheduling.placement.clusters` specifies which clusters the extension **agent components** should be installed to (host, member-4, etc.)

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: openkruise
spec:
  enabled: true
  extension:
    name: openkruise
    version: <exact-version>
  upgradeStrategy: Manual
  # config: |
  #   featureGates: "PreDownloadImageForInPlaceUpdate=true"
  clusterScheduling:
    placement:
      clusters:
      - host
      - member-4
    # overrides:
    #   host: |-
    #     featureGates: "PreDownloadImageForInPlaceUpdate=true"
```

**Fields:**
- `clusterScheduling.placement.clusters`: List of clusters to install the extension **agent** (e.g., `host`, `member-4`). The OpenKruise agent will be deployed to these clusters.
- `clusterScheduling.overrides`: Cluster-specific config overrides (optional)

Apply and verify:

```bash
kubectl apply -f installplan-openkruise.yaml
kubectl get installplan openkruise -w
kubectl describe installplan openkruise
kubectl get extension openkruise -o yaml
```

**Check installation status per cluster:**

```bash
kubectl get installplan openkruise -o jsonpath='{.status.clusterSchedulingStatuses}'
```

### 2.1 Config Handling

`spec.config` is a YAML string block, not a nested object. Only include it when the user asks for non-default settings.

```yaml
spec:
  config: |
    featureGates: "PreDownloadImageForInPlaceUpdate=true"
```

Rules:

- The content under `config: |` must itself be valid YAML
- Do not invent config keys without evidence from chart values, extension documentation, or user-provided requirements
- If config keys are unknown, prefer default install and show discovery commands first

Useful discovery commands:

```bash
kubectl describe extensionversion openkruise-<extension-version>
kubectl get extensionversion openkruise-<extension-version> -o yaml
```

### 3. Status and Logs After Install

```bash
TARGET_NS=$(kubectl get installplan openkruise -o jsonpath='{.status.targetNamespace}')
JOB_NAME=$(kubectl get installplan openkruise -o jsonpath='{.status.jobName}')

kubectl get pods -n "${TARGET_NS:-kruise-system}"
kubectl logs -n "${TARGET_NS:-kruise-system}" -l job-name="$JOB_NAME" --tail=200
kubectl logs -n "${TARGET_NS:-kruise-system}" deploy/kruise-manager --tail=200
kubectl get crd | grep -E 'clonesets.apps.kruise.io|sidecarsets.apps.kruise.io'
```

## Manage SidecarSet

Use `SidecarSet` when the user wants cluster-level sidecar injection and separate lifecycle management for sidecars.

### SidecarSet Template

```yaml
apiVersion: apps.kruise.io/v1alpha1
kind: SidecarSet
metadata:
  name: log-agent
spec:
  selector:
    matchLabels:
      app: demo
  namespaceSelector:
    matchLabels:
      kubernetes.io/metadata.name: demo-project
  containers:
  - name: log-agent
    image: fluent/fluent-bit:2.2
    imagePullPolicy: IfNotPresent
    env:
    - name: LOG_LEVEL
      value: info
    volumeMounts:
    - name: varlog
      mountPath: /var/log
    podInjectPolicy: BeforeAppContainer
  volumes:
  - name: varlog
    hostPath:
      path: /var/log
  updateStrategy:
    type: RollingUpdate
    maxUnavailable: 1
```

Common operations:

```bash
kubectl apply -f sidecarset.yaml
kubectl get sidecarset log-agent -o yaml
kubectl describe sidecarset log-agent
kubectl delete sidecarset log-agent
```

Verify injection:

```bash
kubectl get pods -n demo-project -l app=demo
kubectl get pod <pod-name> -n demo-project -o jsonpath='{.spec.containers[*].name}'
kubectl describe pod <pod-name> -n demo-project
```

SidecarSet troubleshooting checks:

- `spec.selector` matches the workload Pod labels
- `namespaceSelector` matches the target namespace labels
- The sidecar image can be pulled in target namespaces
- `kruise-manager` webhook is healthy

## Manage CloneSet

Use `CloneSet` when the user wants an enhanced workload with in-place image updates, batch rollout, or richer update controls than `Deployment`.

### CloneSet Template

```yaml
apiVersion: apps.kruise.io/v1alpha1
kind: CloneSet
metadata:
  name: sample-app
  namespace: demo-project
spec:
  replicas: 3
  selector:
    matchLabels:
      app: sample-app
  template:
    metadata:
      labels:
        app: sample-app
    spec:
      containers:
      - name: app
        image: nginx:1.25
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 80
  updateStrategy:
    type: InPlaceIfPossible
    partition: 0%
    maxUnavailable: 1
```

Common operations:

```bash
kubectl apply -f cloneset.yaml
kubectl get cloneset sample-app -n demo-project
kubectl describe cloneset sample-app -n demo-project
kubectl delete cloneset sample-app -n demo-project
```

Verify rollout state:

```bash
kubectl get pods -n demo-project -l app=sample-app -o wide
kubectl get events -n demo-project --sort-by=.lastTimestamp
```

### In-place Update Guidance

When the user asks for "in-place update" or an in-place image update:

1. Keep `spec.updateStrategy.type: InPlaceIfPossible`
2. Change the container image in `spec.template.spec.containers`
3. Re-apply the manifest

```bash
kubectl get cloneset sample-app -n demo-project -o yaml > cloneset.yaml
# edit image and, if needed, updateStrategy.partition
kubectl apply -f cloneset.yaml
kubectl describe cloneset sample-app -n demo-project
kubectl get pods -n demo-project -l app=sample-app -o wide
```

### Batch Rollout Guidance

Use `partition` to keep part of the Pods on the old revision during a rollout.

- `partition: 100%` means keep all replicas on the old revision
- `partition: 50%` means only half of the replicas move to the new revision
- `partition: 0%` means finish the rollout

Example rollout sequence:

1. Change image to the new version and set `partition: 100%`
2. Lower to `partition: 50%`
3. Lower to `partition: 0%`

At each step, verify Pod revision movement:

```bash
kubectl get cloneset sample-app -n demo-project -o yaml
kubectl get pods -n demo-project -l app=sample-app -o wide
```

### CloneSet Troubleshooting Checks

- If Pods are recreated instead of in-place updated, confirm `updateStrategy.type` is `InPlaceIfPossible` or `InPlaceOnly`
- Some field changes still require Pod recreation
- If CPU and memory are changed together with image updates in unsupported environments, split the change into separate steps
- Check the controller and workload events with:

```bash
kubectl describe cloneset sample-app -n demo-project
kubectl get pods -n demo-project -l app=sample-app -o wide
kubectl get events -n demo-project --sort-by=.lastTimestamp
```

## Uninstall OpenKruise

### Uninstall Pre-check

Check whether workloads still depend on OpenKruise resources before uninstalling the extension:

```bash
kubectl get sidecarsets.apps.kruise.io -A
kubectl get clonesets.apps.kruise.io -A
```

If these resources still exist, tell the user to migrate or remove them first.

### Default uninstall path:

```bash
kubectl delete installplan openkruise
kubectl get installplan openkruise
kubectl get pods -n kruise-system | grep -i kruise
```

If the user wants full cleanup, remind them to handle application resources first:

- Delete or migrate `SidecarSet` and `CloneSet` resources that are still in use
- Confirm whether application teams still depend on OpenKruise behavior
- Only discuss CRD removal after the user explicitly confirms full cleanup

## Troubleshooting Playbook

### 1. InstallPlan or extension failed

```bash
kubectl describe installplan openkruise
kubectl get installplan openkruise -o jsonpath='{.status.conditions}'
kubectl get extension openkruise -o yaml
kubectl get extensionversion openkruise-<version> -o yaml
```

### 2. Kruise Pods are unhealthy

```bash
kubectl get pods -A | grep -i kruise
kubectl describe pod -n kruise-system <pod-name>
kubectl logs -n kruise-system deploy/kruise-manager --tail=200
kubectl get events -n kruise-system --sort-by=.lastTimestamp
```

### 3. CRD missing or not established

```bash
kubectl get crd clonesets.apps.kruise.io sidecarsets.apps.kruise.io
kubectl describe crd clonesets.apps.kruise.io
kubectl describe crd sidecarsets.apps.kruise.io
kubectl api-resources --api-group apps.kruise.io
```

Likely causes:

- OpenKruise extension installation did not finish
- CRDs were partially created or rejected by the API server
- Webhook or controller startup failed before CRDs became usable

Safe next actions:

- Re-check `InstallPlan` state and conditions
- Inspect install Job logs and `kruise-manager` logs
- Avoid creating workload CRs until the CRDs are established

### 4. Webhook failure

```bash
kubectl get mutatingwebhookconfigurations,validatingwebhookconfigurations | grep -i kruise
kubectl describe mutatingwebhookconfiguration | grep -i -A5 kruise
kubectl describe validatingwebhookconfiguration | grep -i -A5 kruise
kubectl logs -n kruise-system deploy/kruise-manager --tail=200
```

Likely causes:

- `kruise-manager` Pod is not Ready
- Webhook service endpoints are missing
- TLS certificate has expired or does not match the service DNS name
- The API server cannot reach the webhook service
- The target Pod or namespace does not match `SidecarSet` selectors

Safe next actions:

- Check `kruise-manager` Pod readiness and recent restarts
- Check Service and Endpoints for the webhook target
- Inspect recent TLS or x509 errors in controller logs
- Verify the target namespace labels and Pod labels match the `SidecarSet`

### 5. Sidecar not injected

```bash
kubectl get sidecarset -A
kubectl get sidecarset <name> -o yaml
kubectl get pod <pod-name> -n <namespace> -o yaml
kubectl get namespace <namespace> --show-labels
```

### 6. CloneSet rollout stuck

```bash
kubectl describe cloneset <name> -n <namespace>
kubectl get pods -n <namespace> -l app=<label> -o wide
kubectl get events -n <namespace> --sort-by=.lastTimestamp
```

## Output Patterns

Match the answer to the user intent:

- Install request: brief pre-check commands plus a complete `InstallPlan` manifest
- Status request: only the most relevant `kubectl` commands, grouped by purpose
- `SidecarSet` request: one executable manifest plus apply, verify, and delete commands
- `CloneSet` request: one executable manifest plus rollout guidance for in-place update or partition-based rollout
- Uninstall request: delete command, verification commands, and cleanup cautions
- Troubleshooting request: diagnosis commands first, then likely causes, then safe next steps
