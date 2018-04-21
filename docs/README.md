# OpenPitrix Developer Guide

The developer guide is for anyone wanting to either write code which directly accesses the
OpenPitrix API, or to contribute directly to the OpenPitrix project.


## The process of developing and contributing code to the OpenPitrix project

* **Welcome to OpenPitrix (New Developer Guide)**
  ([welcome-to-OpenPitrix-new-developer-guide.md](welcome-to-OpenPitrix-new-developer-guide.md)):
  An introductory guide to contributing to OpenPitrix.

* **On Collaborative Development** ([collab.md](collab.md)): Info on pull requests and code reviews.

* **GitHub Issues** ([issues.md](issues.md)): How incoming issues are triaged.

* **Pull Request Process** ([pull-requests.md](pull-requests.md)): When and why pull requests are closed.

* **Getting Recent Builds** ([getting-builds.md](getting-builds.md)): How to get recent builds including the latest builds that pass CI.

* **Automated Tools** ([automation.md](automation.md)): Descriptions of the automation that is running on our github repository.


## Setting up your dev environment, coding, and debugging

* **Development Guide** ([development.md](development.md)): Setting up your development environment.

* **Testing** ([testing.md](testing.md)): How to run unit, integration, and end-to-end tests in your development sandbox.

* **Hunting flaky tests** ([flaky-tests.md](flaky-tests.md)): We have a goal of 99.9% flake free tests.
  Here's how to run your tests many times.

* **Logging Conventions** ([logging.md](logging.md)): Glog levels.

* **Coding Conventions** ([coding-conventions.md](coding-conventions.md)):
  Coding style advice for contributors.

* **Document Conventions** ([how-to-doc.md](how-to-doc.md))
  Document style advice for contributors.

* **Running a cluster locally** ([running-locally.md](running-locally.md)):
  A fast and lightweight local cluster deployment for development.

## Developing against the OpenPitrix API

* The [REST API documentation](http://openpitrix.io/docs/reference/) explains the REST
  API exposed by apiserver.

## Building releases

See the [openpitrix/release](https://github.com/kubernetes/release) repository for details on creating releases and related tools and helper scripts.
