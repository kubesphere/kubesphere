# Development Workflow

![ks-workflow](docs/images/ks-workflow.png)

## 1 Fork in the cloud

1. Visit https://github.com/kubesphere/kubesphere
2. Click `Fork` button to establish a cloud-based fork.

## 2 Clone fork to local storage

Per Go's [workspace instructions](https://golang.org/doc/code.html#Workspaces), place KubeSphere' code on your `GOPATH` using the following cloning procedure.

1. Define a local working directory:

```bash
$ export working_dir=$GOPATH/src/kubesphere.io
$ export user={your github profile name}
```

2. Create your clone locally:

```bash
$ mkdir -p $working_dir
$ cd $working_dir
$ git clone https://github.com/$user/kubesphere.git
$ cd $working_dir/kubesphere
$ git remote add upstream https://github.com/kubesphere/kubesphere.git

# Never push to upstream master
$ git remote set-url --push upstream no_push

# Confirm that your remotes make sense:
$ git remote -v
```

## 3 Keep your branch in sync

```bash
git fetch upstream
git checkout master
git rebase upstream/master
```

## 4 Add new features or fix issues

Branch from it:

```bash
$ git checkout -b myfeature
```

Then edit code on the myfeature branch.

**Test and build**

Currently, make rules only contain simple checks such as vet, unit test, will add e2e tests soon.

**Using KubeBuilder**

- For Linux OS, you can download and execute this [KubeBuilder script](https://raw.githubusercontent.com/kubesphere/kubesphere/master/hack/install_kubebuilder.sh).

- For MacOS, you can install KubeBuilder by following this [guide](https://book.kubebuilder.io/quick-start.html).

**Run and test**

```bash
$ make all
# Run every unit test
$ make test
```

Run `make help` for additional information on these make targets.

### 5 Development in new branch

**Sync with upstream**

After the test is completed, suggest you to keep your local in sync with upstream which can avoid conflicts.

```
# Rebase your the master branch of your local repo.
$ git checkout master
$ git rebase upstream/master

# Then make your development branch in sync with master branch
git checkout new_feature
git rebase -i master
```
**Commit local changes**

```bash
$ git add <file>
$ git commit -s -m "add your description"
```

### 6 Push to your folk

When ready to review (or just to establish an offsite backup or your work), push your branch to your fork on github.com:

```
$ git push -f ${your_remote_name} myfeature
```

### 7 Create a PR

- Visit your fork at https://github.com/$user/kubesphere
- Click the` Compare & Pull Request` button next to your myfeature branch.
- Check out the [pull request process](pull-request.md) for more details and advice.

