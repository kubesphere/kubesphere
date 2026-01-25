### What type of PR is this?
/kind bug

### What this PR does / why we need it:
This PR addresses the limitation where workspace resource quotas were not correctly supporting extended resources (such as GPUs) unless they were explicitly prefixed with `requests.`.

The `podEvaluator` was filtering out resources that did not appear to be standard resources or prefixed extended resources. This change updates the evaluator to correctly identify and include extended resources in both matching and usage calculation.

### Which issue(s) this PR fixes:
Fixes #4739

### Special notes for reviewers:
The changes encompass:
1.  Modifying `isExtendedResourceNameForQuota` to allow all non-native resources.
2.  Updating `podComputeUsageHelper` to include the original resource name in the usage set for extended resources, ensuring correct matching against quota specifications.

### Does this PR introduced a user-facing change?
```release-note
Fixed an issue where workspace resource quotas did not support extended resources (e.g. GPUs).
```
