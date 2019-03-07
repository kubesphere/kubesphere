# Pull Request Process

This doc explains the process and best practices for submitting a PR to the [KubeSphere project](https://github.com/kubeSphere/kubeSphere). It should serve as a reference for all contributors, and be useful especially to new and infrequent submitters.

- [Before You Submit a PR](#before-you-submit-a-pr)
  * [Run Local Verifications](#run-local-verifications)
- [The PR Submit Process](#the-pr-submit-process)
  * [Write Release Notes if Needed](#write-release-notes-if-needed)
  * [The Testing and Merge Workflow](#the-testing-and-merge-workflow)
  * [Marking Unfinished Pull Requests](#marking-unfinished-pull-requests)
  * [Comment Commands Reference](#comment-commands-reference)
  * [Automation](#automation)
  * [How the e2e Tests Work](#how-the-e2e-tests-work)
- [Why was my PR closed?](#why-was-my-pr-closed)
- [Why is my PR not getting reviewed?](#why-is-my-pr-not-getting-reviewed)
- [Best Practices for Faster Reviews](#best-practices-for-faster-reviews)
  * [0. Familiarize yourself with project conventions](#0-familiarize-yourself-with-project-conventions)
  * [1. Is the feature wanted? Make a Design Doc or Sketch PR](#1-is-the-feature-wanted-make-a-design-doc-or-sketch-pr)
  * [2. Smaller Is Better: Small Commits, Small PRs](#2-smaller-is-better-small-commits-small-prs)
  * [3. Open a Different PR for Fixes and Generic Features](#3-open-a-different-pr-for-fixes-and-generic-features)
  * [4. Comments Matter](#4-comments-matter)
  * [5. Test](#5-test)
  * [6. Squashing and Commit Titles](#6-squashing-and-commit-titles)
  * [7. KISS, YAGNI, MVP, etc.](#7-kiss-yagni-mvp-etc)
  * [8. It's OK to Push Back](#8-its-ok-to-push-back)
  * [9. Common Sense and Courtesy](#9-common-sense-and-courtesy)


# Before You Submit a PR

This guide is for contributors who already have a PR to submit. If you're looking for information on setting up your developer environment and creating code to contribute to KubeSphere, see the [development guide](development.md).

**Make sure your PR adheres to our best practices. These include following project conventions, making small PRs, and commenting thoroughly. Please read the more detailed section on [Best Practices for Faster Reviews](#best-practices-for-faster-reviews) at the end of this doc.**

## Run Local Verifications

You can run these local verifications before you submit your PR to predict the
pass or fail of continuous integration.

# The PR Submit Process

Merging a PR requires the following steps to be completed before the PR will be merged automatically. For details about each step, see the [The Testing and Merge Workflow](#the-testing-and-merge-workflow) section below.

- Make the PR
- Release notes - do one of the following:
  - Add notes in the release notes block, or
  - Update the release note label
- Pass all tests
- Get a `/lgtm` from a reviewer
- Get approval from an owner

If your PR meets all of the steps above, it will enter the submit queue to be merged. When it is next in line to be merged, the tests will run a second time. If all tests pass, the PR will be merged automatically.

## Write Release Notes if Needed

Release notes are required for any PR with user-visible changes, such as bug-fixes, feature additions, and output format changes.

If you don't add release notes in the PR template, the `do-not-merge/release-note-label-needed` label is added to your PR automatically after you create it. There are a few ways to update it.

To add a release-note section to the PR description:

For PRs with a release note:

    ```release-note
    Your release note here
    ```

For PRs that require additional action from users switching to the new release, include the string "action required" (case insensitive) in the release note:

    ```release-note
    action required: your release note here
    ```

For PRs that don't need to be mentioned at release time, just write "NONE" (case insensitive):

    ```release-note
    NONE
    ```

The `/release-note-none` comment command can still be used as an alternative to writing "NONE" in the release-note block if it is left empty.

To see how to format your release notes, view the [PR template](https://github.com/) for a brief example. PR titles and body comments can be modified at any time prior to the release to make them friendly for release notes.
// PR template TODO

Release notes apply to PRs on the master branch. For cherry-pick PRs, see the [cherry-pick instructions](cherry-picks.md). The only exception to these rules is when a PR is not a cherry-pick and is targeted directly to the non-master branch.  In this case, a `release-note-*` label is required for that non-master PR.

// cherry-pick TODO

Now that your release notes are in shape, let's look at how the PR gets tested and merged.

## The Testing and Merge Workflow

The KubeSphere merge workflow uses comments to run tests and labels for merging PRs.
NOTE: For pull requests that are in progress but not ready for review,
prefix the PR title with `WIP` or `[WIP]` and track any remaining TODOs
in a checklist in the pull request description.

Here's the process the PR goes through on its way from submission to merging:

1. Make the pull request
1. `@o8x-merge-robot` assigns reviewers //TODO

If you're **not** a member of the KubeSphere organization:

1. Reviewer/KubeSphere Member checks that the PR is safe to test. If so, they comment `/ok-to-test`
1. Reviewer suggests edits
1. Push edits to your PR branch
1. Repeat the prior two steps as needed
1. (Optional) Some reviewers prefer that you squash commits at this step
1. Owner is assigned and will add the `/approve` label to the PR

If you are a member, or a member comments `/ok-to-test`, the PR will be considered to be trusted. Then the pre-submit tests will run:

1. Automatic tests run. See the current list of tests on the
1. If tests fail, resolve issues by pushing edits to your PR branch
1. If the failure is a flake, anyone on trusted PRs can comment `/retest` to rerun failed tests

Once the tests pass, all failures are commented as flakes, or the reviewer adds the labels `lgtm` and `approved`, the PR enters the final merge queue. The merge queue is needed to make sure no incompatible changes have been introduced by other PRs since the tests were last run on your PR.

Either the [on call contributor](on-call-rotations.md) will manage the merge queue manually. //TODO

1. The PR enters the merge queue ([http://submit-queue.KubeSphere.io]())
1. The merge queue triggers a test re-run with the comment `/test all [submit-queue is verifying that this PR is safe to merge]`
    1. Author has signed the CLA (`cncf-cla: yes` label added to PR)
    1. No changes made since last `lgtm` label applied
1. If tests fail, resolve issues by pushing edits to your PR branch
1. If the failure is a flake, anyone can comment `/retest` if the PR is trusted
1. If tests pass, the merge queue automatically merges the PR

That's the last step. Your PR is now merged.

## Marking Unfinished Pull Requests

If you want to solicit reviews before the implementation of your pull request is complete, you should hold your pull request to ensure that the merge queue does not pick it up and attempt to merge it. There are two methods to achieve this:

1. You may add the `/hold` or `/hold cancel` comment commands
2. You may add or remove a `WIP` or `[WIP]` prefix to your pull request title

The GitHub robots will add and remove the `do-not-merge/hold` label as you use the comment commands and the `do-not-merge/work-in-progress` label as you edit your title. While either label is present, your pull request will not be considered for merging.


## Comment Commands Reference//TODO

[The commands doc]() contains a reference for all comment commands. //TODO

## Automation//TODO

The KubeSphere developer community uses a variety of automation to manage pull requests.  This automation is described in detail [in the automation doc](automation.md). //TODO

## How the Tests Work//TODO

The  tests will post the status results to the PR. If an e2e test fails,
`@o8x-ci-robot` will comment on the PR with the test history and the
comment-command to re-run that test. e.g.

> The following tests failed, say /retest to rerun them all.

# Why was my PR closed?

Pull requests older than 90 days will be closed. Exceptions can be made for PRs that have active review comments, or that are awaiting other dependent PRs. Closed pull requests are easy to recreate, and little work is lost by closing a pull request that subsequently needs to be reopened. We want to limit the total number of PRs in flight to:
* Maintain a clean project
* Remove old PRs that would be difficult to rebase as the underlying code has changed over time
* Encourage code velocity

# Why is my PR not getting reviewed?

A few factors affect how long your PR might wait for review.

If it's the last few weeks of a milestone, we need to reduce churn and stabilize.

Or, it could be related to best practices. One common issue is that the PR is too big to review. Let's say you've touched 39 files and have 8657 insertions. When your would-be reviewers pull up the diffs, they run away - this PR is going to take 4 hours to review and they don't have 4 hours right now. They'll get to it later, just as soon as they have more free time (ha!).

There is a detailed rundown of best practices, including how to avoid too-lengthy PRs, in the next section.

But, if you've already followed the best practices and you still aren't getting any PR love, here are some
things you can do to move the process along:

   * Make sure that your PR has an assigned reviewer (assignee in GitHub). If not, reply to the PR comment stream asking for a reviewer to be assigned.

   * Ping the assignee (@username) on the PR comment stream, and ask for an estimate of when they can get to the review.

   * Ping the assignee on [Slack](http://KubeSphere.slack.com). Remember that a person's GitHub username might not be the same as their Slack username.

   * Ping the assignee by email (many of us have publicly available email addresses).

   * If you're a member of the organization ping the [team](https://github.com/orgs/KubeSphere/teams) (via @team-name) that works in the area you're submitting code.

   * If you have fixed all the issues from a review, and you haven't heard back, you should ping the assignee on the comment stream with a "please take another look" (`PTAL`) or similar comment indicating that you are ready for another review.

Read on to learn more about how to get faster reviews by following best practices.

# Best Practices for Faster Reviews

Most of this section is not specific to KubeSphere, but it's good to keep these best practices in mind when you're making a PR.

You've just had a brilliant idea on how to make KubeSphere better. Let's call that idea Feature-X. Feature-X is not even that complicated. You have a pretty good idea of how to implement it. You jump in and implement it, fixing a bunch of stuff along the way. You send your PR - this is awesome! And it sits. And sits. A week goes by and nobody reviews it. Finally, someone offers a few comments, which you fix up and wait for more review. And you wait. Another week or two go by. This is horrible.

Let's talk about best practices so your PR gets reviewed quickly.

## 0. Familiarize yourself with project conventions

* [Development guide](code-of-conduct.md)

## 1. Is the feature wanted? Make a Design Doc or Sketch PR

Are you sure Feature-X is something the KubeSphere team wants or will accept? Is it implemented to fit with other changes in flight? Are you willing to bet a few days or weeks of work on it?

It's better to get confirmation beforehand. There are two ways to do this:

- Make a proposal doc (in docs/proposals; for example [the QoS proposal]()
- Coordinate your effort with [SIG Docs]() ahead of time. //TODO
- Make a sketch PR (e.g., just the API or Go interface). Write or code up just enough to express the idea and the design and why you made those choices

Or, do all of the above.

Be clear about what type of feedback you are asking for when you submit a proposal doc or sketch PR.

Now, if we ask you to change the design, you won't have to re-write it all.

## 2. Smaller Is Better: Small Commits, Small PRs

Small commits and small PRs get reviewed faster and are more likely to be correct than big ones.

Attention is a scarce resource. If your PR takes 60 minutes to review, the reviewer's eye for detail is not as keen in the last 30 minutes as it was in the first. It might not get reviewed at all if it requires a large continuous block of time from the reviewer.

**Breaking up commits**

Break up your PR into multiple commits, at logical break points.

Making a series of discrete commits is a powerful way to express the evolution of an idea or the
different ideas that make up a single feature. Strive to group logically distinct ideas into separate commits.

For example, if you found that Feature-X needed some prefactoring to fit in, make a commit that JUST does that prefactoring. Then make a new commit for Feature-X.

Strike a balance with the number of commits. A PR with 25 commits is still very cumbersome to review, so use
judgment.

**Breaking up PRs**

Or, going back to our prefactoring example, you could also fork a new branch, do the prefactoring there and send a PR for that. If you can extract whole ideas from your PR and send those as PRs of their own, you can avoid the painful problem of continually rebasing.

KubeSphere is a fast-moving codebase - lock in your changes ASAP with your small PR, and make merges be someone else's problem.

Multiple small PRs are often better than multiple commits. Don't worry about flooding us with PRs. We'd rather have 100 small, obvious PRs than 10 unreviewable monoliths.

We want every PR to be useful on its own, so use your best judgment on what should be a PR vs. a commit.

As a rule of thumb, if your PR is directly related to Feature-X and nothing else, it should probably be part of the Feature-X PR. If you can explain why you are doing seemingly no-op work ("it makes the Feature-X change easier, I promise") we'll probably be OK with it. If you can imagine someone finding value independently of Feature-X, try it as a PR. (Do not link pull requests by `#` in a commit description, because GitHub creates lots of spam. Instead, reference other PRs via the PR your commit is in.)

## 3. Open a Different PR for Fixes and Generic Features

**Put changes that are unrelated to your feature into a different PR.**

Often, as you are implementing Feature-X, you will find bad comments, poorly named functions, bad structure, weak type-safety, etc.

You absolutely should fix those things (or at least file issues, please) - but not in the same PR as your feature. Otherwise, your diff will have way too many changes, and your reviewer won't see the forest for the trees.

**Look for opportunities to pull out generic features.**

For example, if you find yourself touching a lot of modules, think about the dependencies you are introducing between packages. Can some of what you're doing be made more generic and moved up and out of the Feature-X package? Do you need to use a function or type from an otherwise unrelated package? If so, promote! We have places for hosting more generic code.

Likewise, if Feature-X is similar in form to Feature-W which was checked in last month, and you're duplicating some tricky stuff from Feature-W, consider prefactoring the core logic out and using it in both Feature-W and
Feature-X. (Do that in its own commit or PR, please.)

## 4. Comments Matter

In your code, if someone might not understand why you did something (or you won't remember why later), comment it. Many code-review comments are about this exact issue.

If you think there's something pretty obvious that we could follow up on, add a TODO.

Read up on [GoDoc](https://blog.golang.org/godoc-documenting-go-code) - follow those general rules for comments.

## 5. Test

Nothing is more frustrating than starting a review, only to find that the tests are inadequate or absent. Very few PRs can touch code and NOT touch tests.

If you don't know how to test Feature-X, please ask!  We'll be happy to help you design things for easy testing or to suggest appropriate test cases.

## 6. Squashing and Commit Titles

Your reviewer has finally sent you feedback on Feature-X.

Make the fixups, and don't squash yet. Put them in a new commit, and re-push. That way your reviewer can look at the new commit on its own, which is much faster than starting over.

We might still ask you to clean up your commits at the very end for the sake of a more readable history, but don't do this until asked: typically at the point where the PR would otherwise be tagged `LGTM`.

Each commit should have a good title line (<70 characters) and include an additional description paragraph describing in more detail the change intended.

**General squashing guidelines:**

* Sausage => squash

 Do squash when there are several commits to fix bugs in the original commit(s), address reviewer feedback, etc. Really we only want to see the end state and commit message for the whole PR.

* Layers => don't squash

 Don't squash when there are independent changes layered to achieve a single goal. For instance, writing a code munger could be one commit, applying it could be another, and adding a precommit check could be a third. One could argue they should be separate PRs, but there's really no way to test/review the munger without seeing it applied, and there needs to be a precommit check to ensure the munged output doesn't immediately get out of date.

A commit, as much as possible, should be a single logical change.

## 7. KISS, YAGNI, MVP, etc.

Sometimes we need to remind each other of core tenets of software design - Keep It Simple, You Aren't Gonna Need It, Minimum Viable Product, and so on. Adding a feature "because we might need it later" is antithetical to software that ships. Add the things you need NOW and (ideally) leave room for things you might need later - but don't implement them now.

## 8. It's OK to Push Back

Sometimes reviewers make mistakes. It's OK to push back on changes your reviewer requested. If you have a good reason for doing something a certain way, you are absolutely allowed to debate the merits of a requested change. Both the reviewer and reviewee should strive to discuss these issues in a polite and respectful manner.

You might be overruled, but you might also prevail. We're pretty reasonable people. Mostly.

Another phenomenon of open-source projects (where anyone can comment on any issue) is the dog-pile - your PR gets so many comments from so many people it becomes hard to follow. In this situation, you can ask the primary reviewer (assignee) whether they want you to fork a new PR to clear out all the comments. You don't HAVE to fix every issue raised by every person who feels like commenting, but you should answer reasonable comments with an explanation.

## 9. Common Sense and Courtesy

No document can take the place of common sense and good taste. Use your best judgment, while you put
a bit of thought into how your work can be made easier to review. If you do these things your PRs will get merged with less friction.
