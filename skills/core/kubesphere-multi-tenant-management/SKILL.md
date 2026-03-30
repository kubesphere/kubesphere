---
name: kubesphere-multi-tenant-management
description: KubeSphere multi-tenant management Skill. Use when user requests to create users, workspaces, projects, or assign roles/permissions. Supports user lifecycle management, workspace configuration, project creation, role binding. Do not perform any delete operations, do not create custom roles.
---

# KubeSphere Multi-Tenant Management

## Security Guidelines

1. **Never use kubectl edit/delete** - Do NOT use `kubectl edit`, `kubectl delete`, or similar commands to modify or delete workspaces, projects, users, roles, or role bindings. These operations are sensitive and should be performed via KubeSphere Console with proper approval workflow.

2. **Never perform delete operations via API** - Do NOT delete users, workspaces, projects, roles, or role bindings via API. These operations must be performed manually via KubeSphere Console with proper approval workflow. Only use this skill for creating and querying resources.

3. **Never create custom roles** - Do NOT create custom roles (Role, WorkspaceRole, GlobalRole). Only use built-in roles provided by KubeSphere. If custom permissions are needed, instruct the user to configure them via KubeSphere Console.

4. **Default to least privilege** - When creating users or assigning permissions, always use the minimum required access level:
   - New user creation: default to `platform-regular` (not platform-admin)
   - Inviting user to workspace: default to `<workspace-name>-regular` (not admin)
   - Inviting user to project: default to `viewer` (not admin)
   - Only escalate permissions when explicitly requested


## Core Concepts

### Workspace

The top-level organizational unit in KubeSphere, representing a team, department, or business unit. A workspace can contain multiple projects and serves as the basic boundary for resource grouping and access control. **Workspaces can span multiple clusters**, enabling centralized management of resources distributed across different clusters.

### Project

KubeSphere's enhanced Kubernetes namespace, representing a specific application, environment, or workload within a workspace. Each project maps to a separate namespace.

### User & Role

- **User**: KubeSphere account entity, can be platform admin, workspace member, or project member
- **Role**: Permission set defined in KubeSphere's three-tier RBAC:

**Project Roles** (`roles.iam.kubesphere.io`):
- `admin`: Full access to all resources
- `operator`: Create/update/delete resources, cannot manage roles
- `viewer`: Read-only access

**Workspace Roles** (WorkspaceRole, `workspaceroles.iam.kubesphere.io`):
- `<workspace-name>-admin`: Full access to workspace and all projects
- `<workspace-name>-regular`: Limited workspace access
- `<workspace-name>-self-provisioner`: Create projects in workspace
- `<workspace-name>-viewer`: Read-only access to workspace

**Platform Roles** (GlobalRole, `globalroles.iam.kubesphere.io`):
- `platform-admin`: Full access to all resources
- `platform-regular`: Limited platform access
- `platform-self-provisioner`: Can create workspaces

**Role Binding** (KubeSphere API endpoints, binds roles to Users):
- Project-level: `/namespacemembers` API, binds `roles.iam.kubesphere.io` to User
- Workspace-level: `/workspacemembers` API, binds `workspaceroles.iam.kubesphere.io` to User
- Platform-level:  `/users/<username>` API, binds `globalroles.iam.kubesphere.io` to User via annotation

## Step-by-Step Guide

### Prerequisites

Set up authentication using the provided CLI tool. First, navigate to the scripts directory:

```bash
# Navigate to the skill's scripts directory
# Example path (replace with your actual kubesphere-skills location):
cd ~/kubesphere-skills/core/kubesphere-core/scripts

# Install required Python package
pip install requests


# Set host endpoint (optional, defaults to http://ks-apiserver.kubesphere-system)
export KUBESPHERE_HOST="http://<kubesphere-host>"

# Login to get token (token will be cached)
python ks_api.py --login --username admin --password <your-password>

# Token is cached in ~/.kubesphere_token and auto-refreshed

# Optional: Clear cached token
python ks_api.py --clear-cache
```

### 1. Create Workspace

**Required parameters:**
- `workspace-name`: Name for the workspace (maps to `metadata.name`)
- `manager`: Workspace manager (maps to `spec.template.spec.manager`, default to current login user)
- `creator`: Creator name (maps to `metadata.annotations["kubesphere.io/creator"]`)
- `clusters`: List of cluster names to host this workspace (maps to `spec.placement.clusters`)

```bash
# Create workspace via Python CLI
python ks_api.py POST /kapis/tenant.kubesphere.io/v1beta1/workspacetemplates '{
  "apiVersion": "iam.kubesphere.io/v1beta1",
  "kind": "WorkspaceTemplate",
  "metadata": {
    "name": "<workspace-name>",
    "annotations": {
      "kubesphere.io/creator": "<creator>"
    }
  },
  "spec": {
    "template": {
      "spec": {
        "manager": "<manager>"
      },
      "metadata": {
        "annotations": {
          "kubesphere.io/creator": "<creator>"
        }
      }
    },
    "placement": {
      "clusters": [
        {"name": "<cluster-name>"}
      ]
    }
  }
}'
```

**Note:** Before creating a workspace, always ask the user for:
- Workspace name (required)
- Manager (required, default to current login user)
- Clusters (required) - which cluster(s) to assign the workspace to


### 2. Create Project within Workspace

**Required parameters:**
- `project-name`: Name for the project (maps to `metadata.name`)
- `workspace-name`: Name of the workspace to create the project in (maps to `metadata.labels["kubesphere.io/workspace"]`)
- `cluster-name`: Cluster name to create the project in (maps to URI path and `cluster` field)
- `creator`: Creator name (maps to `metadata.annotations["kubesphere.io/creator"]`)

```bash
# Create project within workspace via Python CLI
python ks_api.py POST /clusters/<cluster-name>/kapis/tenant.kubesphere.io/v1beta1/workspaces/<workspace-name>/namespaces '{
  "apiVersion": "v1",
  "kind": "Namespace",
  "metadata": {
    "labels": {
      "kubesphere.io/workspace": "<workspace-name>",
      "kubesphere.io/managed": "true"
    },
    "name": "<project-name>",
    "annotations": {
      "kubesphere.io/creator": "<creator>"
    }
  },
  "cluster": "<cluster-name>"
}'
```

**Note:** Before creating a project, always ask the user for:
- Project name (required)
- Workspace name (required) - which workspace to create the project in
- Cluster name (required) - which cluster to create the project in

### 3. Create User

**Required parameters:**
- `username`: Username for the new user
- `email`: User's email address
- `password`: User's password (must meet KubeSphere password policy)

**Optional parameters:**
- `globalrole`: Platform role (default: `platform-regular`)

```bash
# Create user via Python CLI
python ks_api.py POST /kapis/iam.kubesphere.io/v1beta1/users '{
  "apiVersion": "iam.kubesphere.io/v1beta1",
  "kind": "User",
  "metadata": {
    "annotations": {
      "iam.kubesphere.io/uninitialized": "true",
      "iam.kubesphere.io/globalrole": "platform-regular",
      "kubesphere.io/creator": "admin"
    },
    "name": "<username>"
  },
  "spec": {
    "email": "<email>",
    "password": "<password>"
  }
}'
```

**Note:** Before creating a user, always ask the user for:
- Username (required)
- Email address (required)
- Platform role: If not specified, default to `platform-regular`

### 4. Invite User to Workspace/Project

**For Workspace invitation:**
- `username`: Username to invite (required)
- `workspace-name`: Target workspace name (required)
- `role`: Workspace role (default: `<workspace-name>-regular`)

**For Project invitation:**
- `username`: Username to invite (required)
- `project-name`: Target project name (required)
- `cluster-name`: Cluster name (required)
- `role`: Project role (default: `viewer`)

```bash
# Invite user to workspace (default role: <workspace-name>-regular)
python ks_api.py POST /kapis/iam.kubesphere.io/v1beta1/workspaces/<workspace-name>/workspacemembers '[{"username":"<username>","roleRef":"<workspace-name>-regular"}]'
```

```bash
# Invite user to project (default role: viewer)
python ks_api.py POST /clusters/<cluster-name>/kapis/iam.kubesphere.io/v1beta1/namespaces/<project-name>/namespacemembers '[{"username":"<username>","roleRef":"viewer"}]'
```

**Note:** Before inviting a user, always ask the user for:
- Username to invite (required)
- Target: workspace or project (required)
- Role: If not specified, default to `<workspace-name>-regular` for workspace or `viewer` for project



### 5. Modify User Permissions

Modify user roles at three levels: platform, workspace, and project.

**For Platform Role (global role):**
- `username`: Username to modify (required)
- `globalrole`: New platform role (required)
- Note: Must first GET the user to get current metadata, then PUT with updated annotation

```bash
# Step 1: Get current user info (required before modification)
python ks_api.py GET /kapis/iam.kubesphere.io/v1beta1/users/<username>

# Step 2: Update global role annotation
python ks_api.py PUT /kapis/iam.kubesphere.io/v1beta1/users/<username> '{
  "apiVersion": "iam.kubesphere.io/v1beta1",
  "kind": "User",
  "metadata": {
    "name": "<username>",
    "annotations": {
      "iam.kubesphere.io/globalrole": "<new-global-role>"
    }
  }
}'
```

**For Workspace Role:**
- `username`: Username to modify (required)
- `workspace-name`: Target workspace name (required)
- `roleRef`: New workspace role (required)

```bash
# Modify user role in workspace
python ks_api.py PUT /kapis/iam.kubesphere.io/v1beta1/workspaces/<workspace-name>/workspacemembers/<username> '{"username":"<username>","roleRef":"<workspace-name>-<role>"}'
```

**For Project Role:**
- `username`: Username to modify (required)
- `project-name`: Target project name (required)
- `cluster-name`: Cluster name (required)
- `roleRef`: New project role (required)

```bash
# Modify user role in project
python ks_api.py PUT /clusters/<cluster-name>/kapis/iam.kubesphere.io/v1beta1/namespaces/<project-name>/namespacemembers/<username> '{"username":"<username>","roleRef":"<role>"}'
```

**Note:** Before modifying permissions, always ask the user for:
- Username to modify (required)
- Scope: platform / workspace / project (required)
- New role: Only use built-in roles provided by KubeSphere


### 6. Query Resources

#### List Workspaces
```bash
python ks_api.py GET /kapis/tenant.kubesphere.io/v1beta1/workspacetemplates
```

#### List Users
```bash
python ks_api.py GET /kapis/iam.kubesphere.io/v1beta1/users
```

#### List Workspace Members
```bash
python ks_api.py GET /kapis/iam.kubesphere.io/v1beta1/workspaces/<workspace-name>/workspacemembers
```

#### List Project Members
```bash
python ks_api.py GET /clusters/<cluster-name>/kapis/iam.kubesphere.io/v1beta1/namespaces/<project-name>/namespacemembers
```

#### List Projects in Workspace
```bash
python ks_api.py GET /clusters/<cluster-name>/kapis/tenant.kubesphere.io/v1beta1/workspaces/<workspace-name>/namespaces
```

#### Get User Details
```bash
python ks_api.py GET /kapis/iam.kubesphere.io/v1beta1/users/<username>
```


## Error Handling

| Error Code | Cause | Solution |
|------------|-------|----------|
| `401 Unauthorized` | Token expired | `python ks_api.py --clear-cache && python ks_api.py --login --username admin --password <password>` |
| `403 Forbidden` | No permission | Use admin account |
| `409 Conflict` | Resource already exists | Use different name |
| `404 Not Found` | Resource not found | Verify name/workspace/cluster is correct |
| `400 Bad Request` | Invalid parameters | Check error message for details (email format, password policy, naming rules) |
| Connection refused/timeout | API unreachable | Verify KUBESPHERE_HOST is correct |

**Debugging:**
- `--quiet` flag for cleaner output: `python ks_api.py GET /users --quiet`
- Check token: `python ks_api.py`


## References

- [KubeSphere Workspace Documentation](https://docs.kubesphere.com.cn/v4.2.1/08-workspace-management/)
- [Multi-Tenant Architecture](./references/multi-tenancy-in-kubesphere.md)
- [Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)

## Related Skills

- `kubesphere-core` - Core platform architecture
- `kubesphere-cluster-management` - Cluster operations
