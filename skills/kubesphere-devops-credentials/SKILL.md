---
name: kubesphere-devops-credentials
description: Use when managing credentials in KubeSphere DevOps, including repository credentials, kubeconfig, and API tokens
---

# KubeSphere DevOps Credentials

## Overview

Credentials in KubeSphere DevOps are Kubernetes Secrets with specific labels and annotations. They are synced to Jenkins for use in pipelines. Supported types include SSH keys, username/password, and secret tokens.

## When to Use

- Creating credentials for Git repositories
- Setting up deployment credentials (kubeconfig, registry)
- Managing API tokens for external services
- Troubleshooting credential access issues
- Migrating credentials between DevOps projects

## Credential Types

| Type | Use Case | Secret Key |
|------|----------|------------|
| **SSH** | Git repositories | `username`, `privatekey` |
| **Basic** | Username/password | `username`, `password` |
| **Secret** | API tokens, secrets | `secret` |
| **Kubeconfig** | Kubernetes clusters | `kubeconfig` (v1.1.x only) |
| **SSH Username/Pass** | Git with user/pass | `username`, `password` |
| **String** | Generic text/tokens | `secret` |

## Resource Structure

Credentials are stored as Kubernetes Secrets with DevOps labels:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-credential
  namespace: project-xxx  # DevOps project namespace
  labels:
    devops.kubesphere.io/credential: "true"
  annotations:
    credential.devops.kubesphere.io/syncstatus: successful
    credential.devops.kubesphere.io/type: ssh|basic-auth|secret-text
stringData:
  username: git-user
  privatekey: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    ...
    -----END OPENSSH PRIVATE KEY-----
type: credential.devops.kubesphere.io/ssh  # CRITICAL: Must use credential.devops.kubesphere.io/* type, NOT Opaque!
```


**⚠️ CRITICAL: Secret Type Must Be `credential.devops.kubesphere.io/*`**

The `type` field must be one of:
- `credential.devops.kubesphere.io/basic-auth`
- `credential.devops.kubesphere.io/ssh-auth`
- `credential.devops.kubesphere.io/secret-text`
- `credential.devops.kubesphere.io/kubeconfig`

**Using `type: Opaque` will result in:**
- Credential sync status stuck at "pending"
- Jenkins cannot find the credential
- Pipeline builds fail with "CredentialId could not be found"

**Controller Logic:** The credential controller only watches secrets with types starting with `credential.devops.kubesphere.io/` (see `devopscredential_controller.go` line 102). Secrets with `type: Opaque` are completely ignored.

## API Endpoints

| Operation | Method | Endpoint |
|-----------|--------|----------|
| List Credentials | GET | `/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/credentials` |
| Create Credential | POST | `/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/credentials` |
| Get Credential | GET | `/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/credentials/{credential}` |
| Update Credential | PUT | `/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/credentials/{credential}` |
| Delete Credential | DELETE | `/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/credentials/{credential}` |
| Get Usage | GET | `/kapis/devops.kubesphere.io/v1alpha2/namespaces/{devops}/credentials/{credential}/usage` |

## Common Operations

### List Credentials

```bash
curl "https://kubesphere-api/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/credentials" \
  -H "Authorization: Bearer $TOKEN"
```

### Create SSH Credential

```bash
curl -X POST "https://kubesphere-api/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/credentials" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
      "name": "github-ssh-key",
      "annotations": {
        "credential.devops.kubesphere.io/type": "ssh"
      }
    },
    "stringData": {
      "username": "git",
      "privatekey": "-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----"
    },
    "type": "Opaque"
  }'
```

### Create Basic Auth Credential

```bash
curl -X POST "https://kubesphere-api/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/credentials" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
      "name": "docker-registry",
      "annotations": {
        "credential.devops.kubesphere.io/type": "basic-auth"
      }
    },
    "stringData": {
      "username": "docker-user",
      "password": "docker-password"
    },
    "type": "Opaque"
  }'
```


### Create Basic Auth for Git Access Token (GitHub/GitLab)

**Best Practice:** Use `basic-auth` type for Git access tokens:

```bash
# For GitHub/GitLab access tokens
curl -X POST "https://kubesphere-api/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/credentials" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
      "name": "github-token",
      "annotations": {
        "credential.devops.kubesphere.io/type": "basic-auth"
      }
    },
    "stringData": {
      "username": "git",           # Can be any value for token auth
      "password": "ghp_xxxxxxxxxx"  # Your GitHub/GitLab access token
    },
    "type": "credential.devops.kubesphere.io/basic-auth"
  }'
```

**Why basic-auth for tokens?**
- Git access tokens are used like passwords in HTTPS Git URLs
- ArgoCD and Jenkins both support basic-auth for Git authentication
- Username can be any value (often 'git' or your username)
- Password field holds the actual token

**Supported Git Providers:**
- GitHub Personal Access Token: `ghp_xxxxxxxxxxxx`
- GitLab Personal Access Token: `glpat-xxxxxxxxxx`
- Bitbucket App Password
- Gitea/Forgejo Access Token

### Create Secret Text Credential

```bash
curl -X POST "https://kubesphere-api/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/credentials" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
      "name": "api-token",
      "annotations": {
        "credential.devops.kubesphere.io/type": "secret-text"
      }
    },
    "stringData": {
      "secret": "my-api-token-value"
    },
    "type": "Opaque"
  }'
```

## Using Credentials in Pipelines

### SSH Key for Git Checkout

```groovy
pipeline {
  agent any
  stages {
    stage('Checkout') {
      steps {
        git credentialsId: 'github-ssh-key', url: 'git@github.com:org/repo.git'
      }
    }
  }
}
```

### WithCredentials Step

```groovy
pipeline {
  agent any
  stages {
    stage('Deploy') {
      steps {
        withCredentials([
          usernamePassword(
            credentialsId: 'docker-registry',
            usernameVariable: 'DOCKER_USER',
            passwordVariable: 'DOCKER_PASS'
          )
        ]) {
          sh 'echo $DOCKER_PASS | docker login -u $DOCKER_USER --password-stdin'
        }
      }
    }
  }
}
```

### Kubeconfig (v1.2.x+ with string type)

```groovy
pipeline {
  agent any
  stages {
    stage('Deploy to K8s') {
      steps {
        withCredentials([string(credentialsId: 'my-kubeconfig', variable: 'KUBECONFIG_DATA')]) {
          sh '''
            printf "%s" "$KUBECONFIG_DATA" > kubeconfig
            kubectl --kubeconfig=kubeconfig apply -f deployment.yaml
          '''
        }
      }
    }
  }
}
```


## GitRepository Resource

GitRepository connects a Git repository with a credential for use in pipelines and ArgoCD applications.

### GitRepository Structure

```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: GitRepository
metadata:
  name: my-repo
  namespace: demo-project
spec:
  url: https://github.com/example/repo.git
  provider: github              # Git provider: github, gitlab, bitbucket, etc.
  secret:                       # Reference to credential secret
    name: github-token
    namespace: demo-project
  description: "Main application repository"
```

**Required Fields:**
- `spec.url`: Repository URL
- `spec.provider`: Git provider type (`github`, `gitlab`, `bitbucket`, `gitea`, etc.)
- `spec.secret.name`: Name of the credential secret
- `spec.secret.namespace`: Namespace of the credential secret

### Create GitRepository

**Via API:**
```bash
curl -X POST "https://kubesphere-api/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/gitrepositories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "devops.kubesphere.io/v1alpha3",
    "kind": "GitRepository",
    "metadata": {
      "name": "demo-jenkinsfiles",
      "namespace": "demo-project"
    },
    "spec": {
      "url": "https://github.com/stoneshi-yunify/argocd-example-apps.git",
      "provider": "github",
      "secret": {
        "name": "github-token",
        "namespace": "demo-project"
      },
      "description": "Demo repository with examples"
    }
  }'
```

**Via kubectl:**
```bash
cat <<EOF | kubectl apply -f -
apiVersion: devops.kubesphere.io/v1alpha3
kind: GitRepository
metadata:
  name: my-application-repo
  namespace: demo-project
spec:
  url: https://github.com/example/my-app.git
  provider: github
  secret:
    name: github-token
    namespace: demo-project
  description: "Application source code"
EOF
```

### List GitRepositories

```bash
# Via API
curl "https://kubesphere-api/kapis/devops.kubesphere.io/v1alpha3/namespaces/{devops}/gitrepositories" \
  -H "Authorization: Bearer $TOKEN" | jq '.items[].metadata.name'

# Via kubectl
kubectl get gitrepositories -n demo-project
```

## Credential + GitRepository Usage Patterns

### Pattern 1: Multi-Branch Pipeline with GitRepository

**Complete workflow for private Git repository:**

```yaml
# Step 1: Create credential for Git access
apiVersion: v1
kind: Secret
metadata:
  name: github-token
  namespace: demo-project
  annotations:
    credential.devops.kubesphere.io/type: basic-auth
stringData:
  username: "git"
  password: "ghp_xxxxxxxxxxxxxxxxxxxx"
type: credential.devops.kubesphere.io/basic-auth
---
# Step 2: Create GitRepository linking repo + credential
apiVersion: devops.kubesphere.io/v1alpha3
kind: GitRepository
metadata:
  name: my-app-repo
  namespace: demo-project
spec:
  url: https://github.com/org/my-app.git
  provider: github
  secret:
    name: github-token
    namespace: demo-project
  description: "Application source code"
---
# Step 3: Create multi-branch pipeline using GitRepository
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  name: my-multibranch-pipeline
  namespace: demo-project
spec:
  type: multi-branch-pipeline
  multi_branch_pipeline:
    name: my-multibranch-pipeline
    source_type: git
    git_source:
      url: https://github.com/org/my-app.git
      credential_id: github-token  # Reference credential directly
      discover_branches: true
    script_path: Jenkinsfile
```

### Pattern 2: ArgoCD Application with GitRepository

```yaml
# Step 1: Create credential (basic-auth for token)
apiVersion: v1
kind: Secret
metadata:
  name: github-token
  namespace: demo-project
  annotations:
    credential.devops.kubesphere.io/type: basic-auth
stringData:
  username: "git"
  password: "ghp_xxxxxxxxxxxxxxxxxxxx"
---
# Step 2: Create GitRepository
apiVersion: devops.kubesphere.io/v1alpha3
kind: GitRepository
metadata:
  name: argo-manifests
  namespace: demo-project
spec:
  url: https://github.com/org/k8s-manifests.git
  credentialId: github-token
---
# Step 3: Create ArgoCD Application referencing the repository
apiVersion: gitops.kubesphere.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: demo-project
spec:
  argoApp:
    spec:
      source:
        repoURL: https://github.com/org/k8s-manifests.git
        targetRevision: HEAD
        path: overlays/production
      destination:
        server: https://kubernetes.default.svc
        namespace: demo-project
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
```

### Pattern 3: Complete Setup Script

```bash
#!/bin/bash
set -e

export KUBESPHERE_API="https://kubesphere-api.example.com"
export API_TOKEN="<tenant-token>"
export DEVOPS_PROJECT="demo-project"

# 1. Create credential for GitHub access
echo "Creating GitHub credential..."
curl -s -X POST "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/credentials" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
      "name": "github-token",
      "annotations": {
        "credential.devops.kubesphere.io/type": "basic-auth"
      }
    },
    "stringData": {
      "username": "git",
      "password": "'${GITHUB_TOKEN}'"
    },
    "type": "Opaque"
  }' | jq -r '.metadata.name'

# 2. Create GitRepository
echo "Creating GitRepository..."
curl -s -X POST "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/gitrepositories" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "devops.kubesphere.io/v1alpha3",
    "kind": "GitRepository",
    "metadata": {
      "name": "my-repo",
      "namespace": "'${DEVOPS_PROJECT}'"
    },
    "spec": {
      "url": "https://github.com/'${GITHUB_OWNER}'/'${GITHUB_REPO}'.git",
      "credentialId": "github-token",
      "description": "Application repository"
    }
  }' | jq -r '.metadata.name'

# 3. Create multi-branch pipeline using the repository
echo "Creating multi-branch pipeline..."
curl -s -X POST "${KUBESPHERE_API}/kapis/devops.kubesphere.io/v1alpha3/namespaces/${DEVOPS_PROJECT}/pipelines" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "devops.kubesphere.io/v1alpha3",
    "kind": "Pipeline",
    "metadata": {
      "name": "my-app-pipeline",
      "namespace": "'${DEVOPS_PROJECT}'"
    },
    "spec": {
      "type": "multi-branch-pipeline",
      "multi_branch_pipeline": {
        "name": "my-app-pipeline",
        "source_type": "git",
        "git_source": {
          "url": "https://github.com/'${GITHUB_OWNER}'/'${GITHUB_REPO}'.git",
          "credential_id": "github-token",
          "discover_branches": true,
          "discover_tags": false
        },
        "script_path": "Jenkinsfile"
      }
    }
  }' | jq -r '.metadata.name'

echo "Setup complete!"
```

## Sync Status Troubleshooting

Credentials are synced from Kubernetes to Jenkins. Check sync status:

```bash
# Check credential annotation
kubectl -n <devops-project> get secret <credential-name> -o jsonpath='{.metadata.annotations.credential\.devops\.kubesphere\.io/syncstatus}'

# Force resync all credentials
kubectl get namespaces -l kubesphere.io/devopsproject,devops.kubesphere.io/managed=true --no-headers | \
  awk '{print $1}' | \
  xargs -I{} kubectl annotate secrets credential.devops.kubesphere.io/syncstatus- --all -n {}
```

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| kubeconfigContent not working (v1.2+) | Use `string` type with `withCredentials` |
| Credential not appearing in Jenkins | Check syncstatus annotation |
| SSH auth fails | Ensure username is correct (usually "git") |
| Secret not found in pipeline | Verify credentialsId matches exactly |

| Using wrong credential type for Git tokens | Use `basic-auth` not `secret-text` for Git access tokens |
| GitRepository credential not found | Ensure credential exists before creating GitRepository |
| Git clone fails with 401/403 | Check token hasn't expired and has required permissions |
| ArgoCD cannot access private repo | Verify GitRepository credentialId matches credential name |
| Cannot delete credential | Check if used by pipelines (use `/usage` endpoint) |

## Breaking Changes in v1.2.x

**Removed:** `kubernetes-cd` plugin and `kubeconfigContent` credential type

**Migration:**
```groovy
// v1.1.x (OLD)
withCredentials([kubeconfigContent(credentialsId: 'my-kubeconfig', variable: 'KUBECONFIG_DATA')]) {
  sh 'kubectl --kubeconfig=kubeconfig get node'
}

// v1.2.x (NEW)
withCredentials([string(credentialsId: 'my-kubeconfig', variable: 'KUBECONFIG_DATA')]) {
  sh 'printf "%s" "$KUBECONFIG_DATA" > kubeconfig && kubectl --kubeconfig=kubeconfig get node'
}
```

## References

- [DevOps README - Credentials](/root/go/src/github.com/kubesphere/kse-extensions/devops/README.md)
- [Jenkins Credentials Plugin](https://plugins.jenkins.io/credentials/)
