<!-- Thanks for sending a pull request! Here are some tips for you:

1. If you want **faster** PR reviews, read how: https://github.com/kubesphere/community/blob/master/developer-guide/development/the-pr-author-guide-to-getting-through-code-review.md
2. In case you want to know how your PR got reviewed, read: https://github.com/kubesphere/community/blob/master/developer-guide/development/code-review-guide.md
3. Here are some coding convetions followed by KubeSphere community: https://github.com/kubesphere/community/blob/master/developer-guide/development/coding-conventions.md
-->

### What type of PR is this?
<!-- 
Add one of the following kinds:
/kind bug
/kind cleanup
/kind documentation
/kind feature
/kind design

Optionally add one or more of the following kinds if applicable:
/kind api-change
/kind deprecation
/kind failing-test
/kind flake
/kind regression
-->
/kind bug

### What this PR does / why we need it:
This PR addresses the limitation where workspace resource quotas were not correctly supporting extended resources (such as GPUs) unless they were explicitly prefixed with `requests.`.

The `podEvaluator` was filtering out resources that did not appear to be standard resources or prefixed extended resources. This change updates the evaluator to correctly identify and include extended resources in both matching and usage calculation.

### Which issue(s) this PR fixes:
<!--
Usage: `Fixes #<issue number>`, or `Fixes (paste link of issue)`.
_If PR is about `failing-tests or flakes`, please post the related issues/tests in a comment and do not use `Fixes`_*
-->
Fixes #4739

### Special notes for reviewers:
```
The changes encompass:
1.  Modifying `isExtendedResourceNameForQuota` to allow all non-native resources.
2.  Updating `podComputeUsageHelper` to include the original resource name in the usage set for extended resources, ensuring correct matching against quota specifications.
```

### Does this PR introduced a user-facing change?
<!--
If no, just write "None" in the release-note block below.
If yes, a release note is required:
Enter your extended release note in the block below. If the PR requires additional action from users switching to the new release, include the string "action required".

For more information on release notes see: https://github.com/kubernetes/community/blob/master/contributors/guide/release-notes.md
-->
```release-note
Fixed an issue where workspace resource quotas did not support extended resources (e.g. GPUs).
```

### Additional documentation, usage docs, etc.:
<!--
This section can be blank if this pull request does not require a release note.
Please use the following format for linking documentation or pass the
section below:
- [KEP]: <link>
- [Usage]: <link>
- [Other doc]: <link>
-->
```docs

```
