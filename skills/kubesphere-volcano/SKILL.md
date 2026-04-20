---
name: kubesphere-volcano
description: KubeSphere Volcano job management Skill. Use when user asks to create, list, update, delete Jobs (Volcano Jobs), manage Queues, create PyTorch/TensorFlow/MPI training jobs, or troubleshoot Volcano scheduling issues in KubeSphere. Includes built-in YAML templates, scheduling policy recommendations, and best practices for resource configuration. Handles both KubeSphere API and kubectl operations.
---

# KubeSphere Volcano Management

**Environment (this KubeSphere instance):**
- KubeSphere: Set `KS_HOST` environment variable (e.g., http://<kubesphere-host>:30880)
- Username: admin (default)
- Password: Set `KS_PASSWORD` environment variable
- Clusters: Run `kubectl get clusters` or `ks_api GET /kapis/cluster.kubesphere.io/v1alpha1/clusters`
- Volcano Extension: Run `kubectl get extension volcano -n kubesphere-system` or check via KubeSphere console

Use this skill for the full Volcano lifecycle in KubeSphere:

- Create, list, update, delete Volcano Jobs
- Manage Queues for job scheduling
- Generate YAML templates for PyTorch, TensorFlow, MPI, and batch jobs
- Troubleshoot job pending, pod creation, and scheduling issues

Out of scope by default:

- Advanced Volcano CRDs not requested by the user, such as `JobFlow`, `JobTemplate`, or `Command`
- Deep scheduler configuration without cluster evidence
- Volcano system installation (assume extension is already installed)

If the user explicitly asks for those, acknowledge that they are Volcano capabilities but treat them as a follow-up task.

## Response Rules

- Prefer executable output: `Job` YAML, `Queue` YAML, kubectl commands, or a short ordered procedure.
- Use `kind: Job` (not VolcanoJob), the correct resource type in this KubeSphere environment.
- Use `kubectl get jobs.batch.volcano.sh` or short name `vcjob`/`vj` to query jobs.
- For uninstall/delete requests, warn that deleting Job resources will terminate running pods.
- Omit optional fields instead of guessing values.

**IMPORTANT**: The resource type is `Job` (short: vcjob, vj), NOT `VolcanoJob`. Use these names in all kubectl commands.

## API Usage

This skill supports two approaches for API operations. See **Discovery Commands** section below for detailed commands.


### API Endpoints

KubeSphere provides two API paths for Volcano Job:

#### Option 1: KubeSphere Extension API (/kapis) - Recommended
Uses `volcanojobs` resource name:

```
# Job CRUD (namespace-scoped)
GET    /kapis/batch.volcano.sh/v1alpha1/namespaces/{namespace}/volcanojobs
POST   /kapis/batch.volcano.sh/v1alpha1/namespaces/{namespace}/volcanojobs
GET    /kapis/batch.volcano.sh/v1alpha1/namespaces/{namespace}/volcanojobs/{name}
DELETE /kapis/batch.volcano.sh/v1alpha1/namespaces/{namespace}/volcanojobs/{name}

# Job CRUD (cluster-scoped)
GET    /kapis/batch.volcano.sh/v1alpha1/volcanojobs
DELETE /kapis/batch.volcano.sh/v1alpha1/volcanojobs

# PodGroup CRUD (namespace-scoped)
GET    /kapis/batch.volcano.sh/v1alpha1/namespaces/{namespace}/podgroups

# Queue CRUD (user-scoped)
GET    /kapis/scheduling.volcano.sh/v1beta1/users/{user}/queues
POST   /kapis/scheduling.volcano.sh/v1beta1/users/{user}/queues
DELETE /kapis/scheduling.volcano.sh/v1beta1/users/{user}/queues/{name}
```

#### Option 2: Kubernetes Native API (/apis)
Uses `jobs` resource name:

```
# Job CRUD (namespace-scoped)
GET    /apis/batch.volcano.sh/v1alpha1/namespaces/{namespace}/jobs
POST   /apis/batch.volcano.sh/v1alpha1/namespaces/{namespace}/jobs
GET    /apis/batch.volcano.sh/v1alpha1/namespaces/{namespace}/jobs/{name}
DELETE /apis/batch.volcano.sh/v1alpha1/namespaces/{namespace}/jobs/{name}

# Job CRUD (cluster-scoped)
GET    /apis/batch.volcano.sh/v1alpha1/jobs
DELETE /apis/batch.volcano.sh/v1alpha1/jobs

# Queue CRUD (cluster-scoped)
GET    /apis/scheduling.volcano.sh/v1beta1/queues
POST   /apis/scheduling.volcano.sh/v1beta1/queues
DELETE /apis/scheduling.volcano.sh/v1beta1/queues/{name}
```

> **Note**: Both paths work. Use `/kapis` for KubeSphere extension API (multi-cluster support), use `/apis` for standard Kubernetes API.

> **Note**: Queue API has two views:
> - `/kapis/.../users/{user}/queues` returns queues visible to that user (user-scoped)
> - `/apis/.../queues` returns all queues in the cluster (cluster-scoped)

```
# Query Parameters
page      - Page number (default: 1)
limit     - Items per page
ascending - Sort direction (default: false)
sortBy    - Sort field (e.g., createTime)

```

## Discovery Commands

This section provides two approaches for querying Volcano status:
1. **KubeSphere API (curl)** - for extension management and multi-cluster queries
2. **kubectl** - for direct Kubernetes resource operations

### Option 1: Using KubeSphere API (curl)

**Environment Variables:**
```bash
export KS_HOST="http://<kubesphere-host>:30880"  # KubeSphere console URL (required)
export KS_USERNAME="admin"                         # Username (default)
export KS_PASSWORD="<password>"                    # Password (optional if KS_TOKEN is set)
export KS_TOKEN="<token>"                          # Pre-generated OAuth token (optional, takes priority)
```

```bash
# Get OAuth token - prefer KS_TOKEN if set, otherwise use password
ks_token() {
  # Use KS_TOKEN if it's set and non-empty
  if [ -n "${KS_TOKEN}" ]; then
    echo "$KS_TOKEN"
    return
  fi
  
  # Fall back to password-based token
  if [ -z "${KS_PASSWORD}" ]; then
    echo "Error: KS_TOKEN or KS_PASSWORD must be set" >&2
    return 1
  fi
  
  curl -s -X POST "${KS_HOST}/oauth/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=password&username=${KS_USERNAME:-admin}&password=$KS_PASSWORD&client_id=kubesphere&client_secret=kubesphere" | jq -r '.access_token'
}

# Make API call (supports multi-cluster with optional cluster parameter)
ks_api() {
  local method=$1
  local path=$2
  local cluster=${3:-host}  # Default to host cluster. Pass 3rd arg for member clusters.
  local body=$4
  local token=$(ks_token)
  
  # Check if token is empty
  if [ -z "$token" ]; then
    echo "Error: Failed to obtain authentication token. Please check KS_TOKEN or KS_PASSWORD." >&2
    return 1
  fi
  
  # Prepend cluster path if not already present and not a user-scope path
  if [[ ! "$path" =~ ^/clusters/ ]] && [[ ! "$path" =~ ^/kapis/scheduling.volcano.sh/v1beta1/users ]]; then
    path="/clusters/${cluster}${path}"
  fi
  
  curl -s -X "$method" \
    -H "Authorization: Bearer $token" \
    -H "Content-Type: application/json" \
    ${body:+-d "$body"} \
    "${KS_HOST}$path"
}
```

**Usage:**
```bash
# Option 1: Use pre-generated token (recommended for automation)
export KS_HOST="http://<kubesphere-host>:30880"
export KS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Option 2: Use password (will fetch token each time)
export KS_HOST="http://<kubesphere-host>:30880"
export KS_PASSWORD="your-password"
```

**Query Commands:**

```bash
# List Jobs in namespace (host cluster)
ks_api GET /kapis/batch.volcano.sh/v1alpha1/namespaces/default/volcanojobs

# List all Jobs (cluster-wide)
ks_api GET /kapis/batch.volcano.sh/v1alpha1/volcanojobs

# List Jobs in member cluster (specify cluster explicitly)
ks_api GET /kapis/batch.volcano.sh/v1alpha1/namespaces/default/volcanojobs member-4

# List Queues (user-scoped, no cluster prefix needed)
ks_api GET /kapis/scheduling.volcano.sh/v1beta1/users/admin/queues

# List PodGroups in namespace
ks_api GET /kapis/batch.volcano.sh/v1alpha1/namespaces/default/podgroups

# Check Volcano extension status
ks_api GET /kapis/kubesphere.io/v1alpha1/extensions/volcano

# List available Volcano extension versions
ks_api GET /kapis/kubesphere.io/v1alpha1/extensionversions | jq '.items[] | select(.metadata.name | contains("volcano"))'

# List clusters
ks_api GET /kapis/cluster.kubesphere.io/v1alpha1/clusters
```

### Option 2: Using kubectl (Direct Cluster Access)

```bash
# List all Volcano Jobs
kubectl get jobs.batch.volcano.sh -A
kubectl get vcjob -A
kubectl get vj -A

# List Jobs in specific namespace
kubectl get vcjob -n <namespace>

# List all Queues
kubectl get queue -A

# List all PodGroups
kubectl get podgroup -A

# Check Volcano CRDs
kubectl get crd | grep volcano

# Check Volcano extension
kubectl get extension volcano
kubectl get extensionversion | grep volcano
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
kubectl get jobs.batch.volcano.sh -A
kubectl get queue -A

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
| Create/apply Job/Queue manifests | kubectl |

## Common Operations

### List Jobs

```bash
# List in specific namespace
ks_api GET /kapis/batch.volcano.sh/v1alpha1/namespaces/default/volcanojobs

# List all (cluster-wide)
ks_api GET /kapis/batch.volcano.sh/v1alpha1/volcanojobs

# With kubectl
kubectl get jobs.batch.volcano.sh -A
kubectl get vcjob -A
kubectl get vj -A
kubectl get vcjob -n <namespace>
```

### Get Job Details

```bash
# Via API
ks_api GET /kapis/batch.volcano.sh/v1alpha1/namespaces/default/volcanojobs/my-job

# Via kubectl
kubectl get vcjob my-job -n <namespace> -o yaml
kubectl describe vcjob my-job -n <namespace>

# View logs (get pod name first, then logs)
kubectl get pods -n <namespace> -l volcano.sh/job-name=my-job
kubectl logs <pod-name> -n <namespace>
kubectl logs -f <pod-name> -n <namespace>  # follow mode
```

### Create Job

```bash
# Via API (POST with JSON body - use heredoc for readability)
read -r -d '' JOB_JSON <<'EOF'
{
  "apiVersion": "batch.volcano.sh/v1alpha1",
  "kind": "Job",
  "metadata": {"name": "my-job", "namespace": "default"},
  "spec": {
    "schedulerName": "volcano",
    "queue": "default",
    "tasks": [{
      "replicas": 1,
      "name": "worker",
      "template": {
        "spec": {
          "containers": [{"name": "job", "image": "busybox", "command": ["echo", "hello"]}],
          "restartPolicy": "Never"
        }
      }
    }]
  }
}
EOF
ks_api POST /kapis/batch.volcano.sh/v1alpha1/namespaces/default/volcanojobs "$JOB_JSON"

# Via kubectl (apply YAML)
kubectl apply -f - <<'EOF'
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: my-job
  namespace: default
spec:
  schedulerName: volcano
  queue: default
  tasks:
    - replicas: 1
      name: worker
      template:
        spec:
          containers:
            - name: job
              image: busybox
              command:
                - echo
                - hello
          restartPolicy: Never
EOF
```

### Delete Job

> ⚠️ **WARNING**: Deleting a Job will terminate all running pods associated with it. This action cannot be undone.

```bash
# Via API
ks_api DELETE /kapis/batch.volcano.sh/v1alpha1/namespaces/default/volcanojobs/my-job

# Via kubectl
kubectl delete vcjob my-job -n <namespace>
```

### List Queues

```bash
# Via API
ks_api GET /kapis/scheduling.volcano.sh/v1beta1/users/admin/queues

# Via kubectl
kubectl get queue -A
```

### Create Queue

```bash
# Via kubectl (using YAML)
kubectl apply -f - <<'EOF'
apiVersion: scheduling.volcano.sh/v1beta1
kind: Queue
metadata:
  name: ml-queue
spec:
  weight: 50
  capability:
    cpu: "16"
    memory: "64Gi"
EOF
```

### Update Queue

```bash
# Via kubectl
kubectl patch queue ml-queue -p '{"spec":{"weight":60}}' --type merge
```

### Delete Queue

```bash
# Via API
ks_api DELETE /kapis/scheduling.volcano.sh/v1beta1/users/admin/queues/ml-queue

# Via kubectl
kubectl delete queue ml-queue
```

## Built-in YAML Templates

> **Prerequisite**: The templates reference `claimName: pvc-name`. Ensure the PVC exists in the namespace before applying the Job. Create a PVC first if needed:

### 1. PyTorch Distributed Training Job

```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: pytorch-distributed-training
  namespace: default
spec:
  schedulerName: volcano
  minAvailable: 3
  queue: default
  tasks:
    - replicas: 1
      name: master
      template:
        metadata:
          labels:
            app: pytorch
            role: master
        spec:
          containers:
            - name: pytorch
              image: pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime
              command:
                - python
                - /workspace/train.py
              env:
                - name: MASTER_ADDR
                  value: "$(HOSTNAME)"
                - name: MASTER_PORT
                  value: "29500"
                - name: WORLD_SIZE
                  value: "3"
                - name: RANK
                  value: "0"
              resources:
                requests:
                  cpu: "2"
                  memory: "4Gi"
                limits:
                  cpu: "2"
                  memory: "4Gi"
              volumeMounts:
                - name: workspace
                  mountPath: /workspace
          volumes:
            - name: workspace
              persistentVolumeClaim:
                claimName: pvc-name
          restartPolicy: Never
    - replicas: 2
      name: worker
      template:
        metadata:
          labels:
            app: pytorch
            role: worker
        spec:
          containers:
            - name: pytorch
              image: pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime
              command:
                - python
                - /workspace/train.py
              env:
                - name: MASTER_ADDR
                  value: pytorch-distributed-training-master-0  # Format: {job}-{task}-{index}
                - name: MASTER_PORT
                  value: "29500"
                - name: WORLD_SIZE
                  value: "3"
                - name: RANK
                  value: "$(VOLCANO_TASK_INDEX)"  # worker: 0,1,2... (global RANK = this + master_replicas)
              resources:
                requests:
                  cpu: "2"
                  memory: "4Gi"
                limits:
                  cpu: "2"
                  memory: "4Gi"
              volumeMounts:
                - name: workspace
                  mountPath: /workspace
          volumes:
            - name: workspace
              persistentVolumeClaim:
                claimName: pvc-name
          restartPolicy: Never
  policies:
    - event: TaskCompleted
      action: CompleteJob
```

### 2. TensorFlow Training Job

```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: tf-training
  namespace: default
spec:
  schedulerName: volcano
  minAvailable: 2
  queue: default
  tasks:
    - replicas: 1
      name: ps
      template:
        metadata:
          labels:
            app: tf
            role: ps
        spec:
          containers:
            - name: tensorflow
              image: tensorflow/tensorflow:2.14.0-gpu
              command:
                - python
                - /workspace/train.py
              resources:
                requests:
                  cpu: "2"
                  memory: "4Gi"
                  nvidia.com/gpu: "1"
                limits:
                  cpu: "2"
                  memory: "4Gi"
                  nvidia.com/gpu: "1"
              volumeMounts:
                - name: workspace
                  mountPath: /workspace
          volumes:
            - name: workspace
              persistentVolumeClaim:
                claimName: pvc-name
          restartPolicy: Never
    - replicas: 1
      name: worker
      template:
        metadata:
          labels:
            app: tf
            role: worker
        spec:
          containers:
            - name: tensorflow
              image: tensorflow/tensorflow:2.14.0-gpu
              command:
                - python
                - /workspace/train.py
              resources:
                requests:
                  cpu: "2"
                  memory: "4Gi"
                  nvidia.com/gpu: "1"
                limits:
                  cpu: "2"
                  memory: "4Gi"
                  nvidia.com/gpu: "1"
              volumeMounts:
                - name: workspace
                  mountPath: /workspace
          volumes:
            - name: workspace
              persistentVolumeClaim:
                claimName: pvc-name
          restartPolicy: Never
```

### 3. MPI Job

```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: mpi-job
  namespace: default
spec:
  schedulerName: volcano
  minAvailable: 3
  queue: default
  tasks:
    - replicas: 1
      name: launcher
      template:
        metadata:
          labels:
            app: mpi
            role: launcher
        spec:
          containers:
            - name: mpi
              image: mpioperator/mpich:latest
              command:
                - mpirun
                - -np
                - "4"
                - ./run.sh
              resources:
                requests:
                  cpu: "1"
                  memory: "2Gi"
                limits:
                  cpu: "1"
                  memory: "2Gi"
          restartPolicy: Never
    - replicas: 2
      name: worker
      template:
        metadata:
          labels:
            app: mpi
            role: worker
        spec:
          containers:
            - name: mpi
              image: mpioperator/mpich:latest
              resources:
                requests:
                  cpu: "2"
                  memory: "4Gi"
                limits:
                  cpu: "2"
                  memory: "4Gi"
          restartPolicy: Never
```

### 4. Simple Batch Job

```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: batch-job
  namespace: default
spec:
  schedulerName: volcano
  minAvailable: 1
  queue: default
  tasks:
    - replicas: 4
      name: worker
      template:
        metadata:
          labels:
            app: batch
        spec:
          containers:
            - name: job
              image: busybox:latest
              command:
                - sh
                - -c
                - |
                  echo "Processing batch data..."
                  sleep 30
                  echo "Done"
              resources:
                requests:
                  cpu: "1"
                  memory: "1Gi"
                limits:
                  cpu: "1"
                  memory: "1Gi"
          restartPolicy: Never
```

### 5. Queue Configuration

```yaml
apiVersion: scheduling.volcano.sh/v1beta1
kind: Queue
metadata:
  name: ml-queue
spec:
  weight: 50
  capability:
    cpu: "16"
    memory: "64Gi"
  # Optional: resource quota (when using hierarchy)
  # quota:
  #   minResource:
  #     cpu: "8"
  #     memory: "32Gi"
  #   maxResource:
  #     cpu: "32"
  #     memory: "128Gi"
```

## Best Practices

### Resource Configuration

| Workload Type | CPU | Memory | GPU | Notes |
|---------------|-----|--------|-----|-------|
| PyTorch Training | 2-4 per replica | 4-8Gi per replica | 1-2 per replica | Use GPU instances |
| TensorFlow Training | 2-4 per replica | 4-8Gi per replica | 1 per replica | Match GPU to model size |
| MPI Job | 1-2 per rank | 2-4Gi per rank | Optional | Minimize network latency |
| Batch Processing | 1-2 per task | 1-4Gi per task | None | Scale horizontally |

### Scheduling Recommendations

1. **Use appropriate queue** - Create queues for different workload priorities:
   - `default` - Regular jobs
   - `ml-queue` - ML training jobs (higher priority)
   - `low-priority` - Batch jobs that can wait

2. **Set proper minAvailable** - Ensure enough pods are ready before job starts:
   - Distributed training: set to `replicas - 1` (allow 1 failure)
   - Batch jobs: set to `1` or `replicas` based on requirements

3. **Configure retry policies** - Add policies for job recovery:
   ```yaml
   policies:
     - event: TaskFailed
       action: RestartTask
     - event: PodEvicted
       action: RestartTask
   ```

4. **Use appropriate restart policy**:
   - `Never` - For distributed jobs where restarts create new pods
   - `OnFailure` - For jobs that can recover from failures

### Troubleshooting

```bash
# Check Job status
kubectl get vcjob <name> -n <namespace>
kubectl describe vcjob <name> -n <namespace>

# Check PodGroup status
kubectl get podgroup <name> -n <namespace> -o yaml

# Check pods created by Job
kubectl get pods -n <namespace> -l volcano.sh/job-name=<name>

# Check volcano system pods
kubectl get pods -n volcano-system

# List all Volcano Jobs
kubectl get jobs.batch.volcano.sh -A

# List Queues
kubectl get queue -A

# Common issues:
# 1. Job pending - check PodGroup status and events
# 2. Pods not created - check scheduler and queue resources
# 3. Pods evicted - check queue capacity and priority
# 4. Job not creating pods - check volcano controller is running in volcano-system namespace
```

## Response Patterns

Match output to user intent:

| Request Type | Output |
|--------------|--------|
| Create job | YAML manifest + kubectl apply command + verification |
| List jobs | kubectl/ks_api command + explanation |
| Get job details | kubectl describe + relevant sections |
| Delete job | kubectl delete command + confirmation |
| Show templates | Template with placeholders + usage notes |
| Troubleshooting | Diagnostic commands first, then causes and solutions |
| Best practices | Context-aware recommendations |

## Workload Type Detection

When user describes a job, infer the type:

| User Says | Use Template |
|-----------|--------------|
| "训练", "training", "train", "机器学习" | PyTorch or TensorFlow |
| "分布式训练", "distributed training" | PyTorch Distributed |
| "MPI", "mpi" | MPI Job |
| "批处理", "batch", "批量任务" | Simple Batch Job |
| "队列", "queue" | Queue Configuration |

Apply the template with appropriate modifications based on user requirements.
