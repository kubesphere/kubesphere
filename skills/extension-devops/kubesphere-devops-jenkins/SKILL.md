---
name: kubesphere-devops-jenkins
description: Use when configuring Jenkins in KubeSphere DevOps, including agent customization, LDAP/OIDC integration, build artifact retrieval, or troubleshooting Jenkins issues
---

# KubeSphere DevOps Jenkins Configuration

## Overview

KubeSphere DevOps embeds Jenkins as the CI engine. Jenkins is configured via Configuration as Code (CasC) and provides the underlying pipeline execution environment. Understanding Jenkins configuration is essential for customizing agents, authentication, and resource management.

## When to Use

- Accessing Jenkins console directly
- Configuring LDAP or OIDC authentication
- Customizing Jenkins agent images
- Configuring GitLab or other SCM servers
- Troubleshooting Jenkins startup issues
- Updating Jenkins after DevOps upgrade
- Triggering builds via API
- Downloading build artifacts
- Viewing build logs and status

## Accessing Jenkins Console

### Get Admin Credentials

```bash
# Get Jenkins admin user and password
kubectl -n kubesphere-devops-system get secret devops-jenkins -o yaml

# Decode values
echo "<jenkins-admin-password-base64>" | base64 -d
echo "<jenkins-admin-user-base64>" | base64 -d
```

### Get Jenkins NodePort

```bash
kubectl -n kubesphere-devops-system get svc devops-jenkins
# Default NodePort: 30180
```

Access via: `http://<master-node-ip>:30180`

## Configuration As Code (CasC)

Jenkins configuration is managed through the `jenkins-casc-config` ConfigMap:

```bash
# View current CasC
kubectl -n kubesphere-devops-system get cm jenkins-casc-config -o yaml
```

### Key Configuration Sections

```yaml
agent:
  jenkins:
    Master:
      NodeSelector: {}
      Tolerations: []
    Agent:
      Image: "jenkins/inbound-agent"
      Tag: "3309.v27b_9314fd1a_4-1-jdk21"
      Privileged: false
      NodeSelector: {}
```

## Authentication Configuration

### LDAP Integration

```yaml
agent:
  jenkins:
    exactSecurityRealm:
      ldap:
        configurations:
        - displayNameAttributeName: "uid"
          mailAddressAttributeName: "mail"
          inhibitInferRootDN: false
          managerDN: "cn=admin,dc=kubesphere,dc=io"
          managerPasswordSecret: "admin"
          rootDN: "dc=kubesphere,dc=io"
          userSearchBase: "ou=Users"
          userSearch: "(&(objectClass=inetOrgPerson)(|(uid={0})(mail={0})))"
          groupSearchBase: "ou=Groups"
          groupSearchFilter: "(&(objectClass=posixGroup)(cn={0}))"
          server: "ldap://openldap.kubesphere-system.svc:389"
        disableMailAddressResolver: false
        disableRolePrefixing: true
```

### OpenID Connect (OIDC) Integration

```yaml
agent:
  jenkins:
    exactSecurityRealm:
      oic:
        clientId: "jenkins"
        clientSecret: "jenkins"
        tokenServerUrl: "http://192.168.1.20:30880/oauth/token"
        authorizationServerUrl: "http://192.168.1.20:30880/oauth/authorize"
        userInfoServerUrl: "http://192.168.1.20:30880/oauth/userinfo"
        endSessionEndpoint: "http://192.168.1.20:30880/oauth/logout"
        logoutFromOpenidProvider: true
        scopes: openid profile email
        fullNameFieldName: url
        userNameField: preferred_username
    redirectURIs:
    - http://192.168.1.20:30180/securityRealm/finishLogin
```

## GitLab Integration

Configure GitLab servers for pipeline integration:

```yaml
agent:
  jenkins:
    unclassified:
      gitLabServers:
        - name: "gitlab-a"
          serverUrl: "https://gitlab.a.com"
        - name: "gitlab-b"
          serverUrl: "https://gitlab.b.com"
```

After updating ConfigMap, create credentials in Jenkins Console:
1. Manage Jenkins > System
2. Find GitLab section
3. Add credentials (GitLab Personal Access Token)
4. Test connection
5. Save

## Agent Customization

### Update JNLP Image (v1.2.x)

```bash
# Get current config
kubectl -n kubesphere-devops-system get cm jenkins-casc-config -o yaml > /tmp/casc-old.yaml

# Update image version
sed 's/inbound-agent:4.10-2/inbound-agent:3309.v27b_9314fd1a_4-1-jdk21/g' /tmp/casc-old.yaml > /tmp/casc.yaml

# Apply new config
kubectl apply -f /tmp/casc.yaml

# Restart Jenkins
kubectl -n kubesphere-devops-system rollout restart deployment devops-jenkins
```

### Enable Podman for Non-Docker Environments

For hosts running containerd instead of Docker:

```yaml
agent:
  jenkins:
    Agent:
      Privileged: true  # Required for podman
```

Use agent images with `-podman` suffix. These alias `docker` command to `podman`.

### Custom Agent PodTemplate

```yaml
agent:
  jenkins:
    Agent:
      PodTemplate:
        Name: "default"
        Label: "jenkins-agent"
        Containers:
        - Name: "jnlp"
          Image: "jenkins/inbound-agent:3309.v27b_9314fd1a_4-1-jdk21"
          Args: "^${computer.jnlpmac} ^${computer.name}"
          Resource:
            Request:
              Cpu: "100m"
              Memory: "256Mi"
            Limit:
              Cpu: "500m"
              Memory: "512Mi"
```

## Backup and Restore

### Backup Jenkins Data

```bash
# Find Jenkins PVC
kubectl -n kubesphere-devops-system get pvc

# Create backup inside Jenkins pod
kubectl -n kubesphere-devops-system exec deployment/devops-jenkins -- bash -c \
  'cd /tmp && tar czvf jenkins_home.backup.tar /var/jenkins_home && mv jenkins_home.backup.tar /var/jenkins_home'

# Copy to local
kubectl -n kubesphere-devops-system cp \
  deployment/devops-jenkins:/var/jenkins_home/jenkins_home.backup.tar \
  ./jenkins_home.backup.tar
```

### Reset Jenkins Components

```yaml
agent:
  jenkins:
    Master:
      resetPlugins: true    # Reset all plugins to default
      resetRBACRoles: true  # Reset RBAC roles
      resetAdminPassword: true  # Reset admin password
      resetAdminToken: true     # Reset admin API token
```

Apply and restart Jenkins to reset.

## Troubleshooting

### Jenkins Won't Start

```bash
# Check pod status
kubectl -n kubesphere-devops-system get pods -l app=devops-jenkins

# View logs
kubectl -n kubesphere-devops-system logs -l app=devops-jenkins --tail=100

# Check events
kubectl -n kubesphere-devops-system get events --sort-by='.lastTimestamp'
```

### Common Issues

| Issue | Cause | Fix |
|-------|-------|-----|
| OOMKilled | Memory limit too low | Increase resource limits |
| CrashLoopBackOff | Bad CasC config | Check ConfigMap syntax |
| Agent connection failed | Wrong JNLP image | Update to compatible version |
| Pipeline hangs | No available agents | Check agent resource quotas |

### Check Jenkins Health

```bash
# Check crumb issuer (CSRF protection)
curl http://<jenkins-url>/crumbIssuer/api/json

# Check plugin list
curl http://<jenkins-url>/pluginManager/api/json?depth=1
```

## Plugin Management

### v1.2.x Removed Plugins

These plugins were removed in v1.2.0:
- `ace-editor`
- `async-http-client`
- `blueocean-executor-info`
- `handlebars`
- `kubernetes-cd` (major impact)
- `momentjs`
- `windows-slaves`

**Action Required:** Update pipeline scripts if they depend on these plugins.

## Working with Builds via API

### Trigger a Pipeline Build

```bash
# Get Jenkins admin token
TOKEN=$(kubectl -n kubesphere-devops-system get secret devops-jenkins -o jsonpath='{.data.jenkins-admin-token}' | base64 -d)

# Trigger build for a pipeline
kubectl run curl-trigger --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/job/my-pipeline/build" \
  -X POST -w "\nHTTP Status: %{http_code}\n"

# Expected: HTTP Status: 201 (Created)
```

**Multi-branch Pipeline:**
```bash
# Trigger specific branch build
kubectl run curl-trigger-branch --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/job/my-multibranch/job/main/build" \
  -X POST
```

### View Build Console Log

```bash
# Get console log for build #3
kubectl run curl-log --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/job/my-pipeline/job/main/3/consoleText"
```

### Check Build Status

```bash
# Get build info as JSON
kubectl run curl-status --rm -i --restart=Never --image=curlimages/curl \
  -- "http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/job/my-pipeline/job/main/3/api/json" \
  | grep -E '"result"|"building"|"duration"'

# Example output:
# "building":false
# "duration":53917
# "result":"SUCCESS"
```

### Download Build Artifacts

**Method: Pod-based Download (Recommended for Binaries)**

For binary artifacts, use a pod to download and then copy out:

```bash
# 1. Create a download pod
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: artifact-downloader
spec:
  containers:
  - name: downloader
    image: curlimages/curl
    command: ['sh', '-c', 'sleep 300']
EOF

# 2. Wait for pod ready
kubectl wait --for=condition=Ready pod/artifact-downloader --timeout=60s

# 3. Get Jenkins token and download artifact
TOKEN=$(kubectl -n kubesphere-devops-system get secret devops-jenkins -o jsonpath='{.data.jenkins-admin-token}' | base64 -d)
kubectl exec artifact-downloader -- sh -c \
  "curl -s -o /tmp/service 'http://admin:${TOKEN}@devops-jenkins.kubesphere-devops-system:80/job/demo-project/job/my-pipeline/job/main/3/artifact/service'"

# 4. Copy artifact from pod to local
kubectl cp artifact-downloader:/tmp/service /tmp/service

# 5. Clean up
kubectl delete pod artifact-downloader --force
```

**Verify Downloaded Artifact:**
```bash
# Check file
ls -lh /tmp/service
file /tmp/service

# Example output:
# /tmp/service: ELF 64-bit LSB executable, x86-64...

# Make executable and test
chmod +x /tmp/service
/tmp/service --help
```

### Working with PipelineRuns

PipelineRuns track Jenkins build execution in Kubernetes:

```bash
# List recent pipeline runs
kubectl get pipelineruns -n <devops-project-namespace> --sort-by=.metadata.creationTimestamp

# Example output:
# NAME                             COMPLETIONS   STATUS      AGE
# my-pipeline-vf8p5                1             Succeeded   2m

# Get detailed status
kubectl get pipelinerun my-pipeline-vf8p5 -n demo-project -o yaml
```

### Common API Patterns

| Task | API Endpoint |
|------|--------------|
| Trigger build | `/job/{folder}/job/{pipeline}/build` |
| Get build status | `/job/{folder}/job/{pipeline}/job/{branch}/{number}/api/json` |
| Get console log | `/job/{folder}/job/{pipeline}/job/{branch}/{number}/consoleText` |
| Get artifact | `/job/{folder}/job/{pipeline}/job/{branch}/{number}/artifact/{filename}` |
| List builds | `/job/{folder}/job/{pipeline}/job/{branch}/api/json` |

**Folder Structure:**
- DevOps Project → Folder in Jenkins (e.g., `demo-project`)
- Pipeline → Job in folder (e.g., `my-pipeline`)
- Branch → Sub-job for multi-branch (e.g., `main`)
- Build Number → Individual run (e.g., `3`)

## API Proxy Endpoints

KubeSphere proxies Jenkins API for authentication:

| Endpoint | Description |
|----------|-------------|
| `/kapis/devops.kubesphere.io/v1alpha2/jenkins/{path}` | Generic Jenkins API proxy |
| `/kapis/devops.kubesphere.io/v1alpha2/namespaces/{devops}/jenkins/{path}` | Project-scoped proxy |
| `/kapis/devops.kubesphere.io/v1alpha3/ci/nodelabels` | Get Jenkins node labels |

## References

- [DevOps README - Configuration](/root/go/src/github.com/kubesphere/kse-extensions/devops/README.md)
- [Jenkins CasC Documentation](https://github.com/jenkinsci/configuration-as-code-plugin)
- [Jenkins Pipeline Syntax](https://www.jenkins.io/doc/book/pipeline/syntax/)
