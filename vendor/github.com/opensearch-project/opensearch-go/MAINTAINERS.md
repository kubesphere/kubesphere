- [Overview](#overview)
- [Current Maintainers](#current-maintainers)
- [Maintainer Responsibilities](#maintainer-responsibilities)
    - [Uphold Code of Conduct](#uphold-code-of-conduct)
    - [Prioritize Security](#prioritize-security)
    - [Review Pull Requests](#review-pull-requests)
    - [Triage Open Issues](#triage-open-issues)
    - [Be Responsive](#be-responsive)
    - [Maintain Overall Health of the Repo](#maintain-overall-health-of-the-repo)
    - [Use Semver](#use-semver)
    - [Release Frequently](#release-frequently)
    - [Promote Other Maintainers](#promote-other-maintainers)

## Overview

This document explains who the maintainers are (see below), what they do in this repo, and how they should be doing it. If you're interested in contributing, see [CONTRIBUTING](CONTRIBUTING.md).

## Current Maintainers

| Maintainer               | GitHub ID                               | Affiliation |
| ------------------------ | --------------------------------------- | ----------- |
| Jack Mazanec             | [jmazanec15](https://github.com/jmazanec15) | Amazon |
| Vamshi Vijay Nakkirtha   | [vamshin](https://github.com/vamshin)   |   Amazon    |
| Vijayan Balasubramanian  | [VijayanB](https://github.com/VijayanB) |   Amazon    |

## Maintainer Responsibilities

Maintainers are active and visible members of the community, and have [maintain-level permissions on a repository](https://docs.github.com/en/organizations/managing-access-to-your-organizations-repositories/repository-permission-levels-for-an-organization). Use those privileges to serve the community and evolve code as follows.

### Uphold Code of Conduct

Model the behavior set forward by the [Code of Conduct](CODE_OF_CONDUCT.md) and raise any violations to other maintainers and admins.

### Prioritize Security

Security is your number one priority. Maintainer's Github keys must be password protected securely and any reported security vulnerabilities are addressed before features or bugs.

Note that this repository is monitored and supported 24/7 by Amazon Security, see [Reporting a Vulnerability](SECURITY.md) for details.

### Review Pull Requests

Review pull requests regularly, comment, suggest, reject, merge and close. Accept only high quality pull-requests. Provide code reviews and guidance on incomming pull requests. Don't let PRs be stale and do your best to be helpful to contributors.

### Triage Open Issues

Manage labels, review issues regularly, and triage by labelling them.

All repositories in this organization have a standard set of labels, including `bug`, `documentation`, `duplicate`, `enhancement`, `good first issue`, `help wanted`, `blocker`, `invalid`, `question`, `wontfix`, and `untriaged`, along with release labels, such as `v1.0.0`, `v1.1.0` and `v2.0.0`, and `backport`.

Use labels to target an issue or a PR for a given release, add `help wanted` to good issues for new community members, and `blocker` for issues that scare you or need immediate attention. Request for more information from a submitter if an issue is not clear. Create new labels as needed by the project.

### Be Responsive

Respond to enhancement requests, and forum posts. Allocate time to reviewing and commenting on issues and conversations as they come in.

### Maintain Overall Health of the Repo

Keep the `main` branch at production quality at all times. Backport features as needed. Cut release branches and tags to enable future patches.

### Use Semver

Use and enforce [semantic versioning](https://semver.org/) and do not let breaking changes be made outside of major releases.

### Release Frequently

Make frequent project releases to the community.

### Promote Other Maintainers

Assist, add, and remove [MAINTAINERS](MAINTAINERS.md). Exercise good judgement, and propose high quality contributors to become co-maintainers.
