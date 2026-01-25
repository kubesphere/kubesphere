# Allow workspace resource quota to support non-prefixed extended resources (e.g. GPU)

**What type of PR is this?**
/kind feature

**What this PR does / why we need it**:
This PR enables support for defining workspace resource quotas for extended resources (such as GPUs) without enforcing the `requests.` prefix. 

Previously, KubeSphere only supported extended resources in quotas if they were prefixed with `requests.` (e.g., `requests.nvidia.com/gpu`). This limitation prevented users from intuitively defining quotas using the direct resource name (e.g., `nvidia.com/gpu`), which is a common requirement for managing hardware accelerators.

This change modifies the `kube/pkg/quota/v1/evaluator/core` package to:
1.  Relax the validation in `isExtendedResourceNameForQuota` to allow extended resources without the `requests.` prefix.
2.  Update `podComputeUsageHelper` to calculate usage for both the prefixed and non-prefixed versions of extended resources.

**Which issue(s) this PR fixes**:
Fixes #4739

**Special notes for your reviewer**:
I have added a unit test `kube/pkg/quota/v1/evaluator/core/pods_test.go` to verify the fix.
- `TestPodEvaluatorMatchingResources`: Verifies that `nvidia.com/gpu` is correctly identified as a matching resource.
- `TestPodComputeUsageHelper`: Verifies that the usage calculation includes the non-prefixed resource name.

**Does this PR introduce a user-facing change?**:
```release-note
feature: Support defining workspace resource quotas for extended resources (e.g., GPUs) without the "requests." prefix.
```
