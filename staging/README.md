# External Repository Staging Area

This directory is the staging area for packages that have been split to their
own repository. The content here will be periodically published to respective
top-level kubesphere.io repositories.

Repositories currently staged here:

- [`kubesphere.io/client-go`](https://github.com/kubesphere/client-go)
- [`kubesphere.io/api`](https://github.com/kubesphere/api)


The code in the staging/ directory is authoritative, i.e. the only copy of the
code. You can directly modify such code.

## Using staged repositories from KubeSphere code

KubeSphere code uses the repositories in this directory via symlinks in the
`vendor/kubesphere.io` directory into this staging area. For example, when
KubeSphere code imports a package from the `kubesphere.io/client-go` repository, that
import is resolved to `staging/src/kubesphere.io/client-go` relative to the project
root:

```go
// pkg/example/some_code.go
package example

import (
  "kubesphere.io/client-go/" // resolves to staging/src/kubesphere.io/client-go/dynamic
)
```

Once the change-over to external repositories is complete, these repositories
will actually be vendored from `kubesphere.io/<package-name>`.

## Creating a new repository in staging

### Adding the staging repository in `kubesphere/kubesphere`:

1. Add a propose to sig-architecture in [community](https://github.com/kubesphere/community/). Waiting approval for creating the staging repository.

2. Once approval has been granted, create the new staging repository.

3. Add a symlink to the staging repo in `vendor/kubesphere.io`.

4. Add all mandatory template files to the staging repo such as README.md, LICENSE, OWNER,CONTRIBUTING.md.


### Creating the published repository

1. Create an repository in the KubeSphere org. The published repository **must** have an
initial empty commit.

2. Setup branch protection and enable access to the `ks-publishing-bot` bot.

3. Once the repository has been created in the KubeSphere org, update the publishing-bot to publish the staging repository by updating:

    - [`rules.yaml`](/staging/publishing/rules.yaml):
    Make sure that the list of dependencies reflects the staging repos in the `Godeps.json` file.

4. Add the repo to the list of staging repos in this `README.md` file.