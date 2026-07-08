# Substrate GitHub App Setup Guide

This guide walks you through installing the Substrate GitHub App and enabling its merge-blocking functionality to protect your data contracts.

## Installation Steps

1. **Install the App**: Navigate to the [Substrate GitHub App](https://github.com/apps/substrate) in the GitHub Marketplace (or use your private installation link).
2. **Select Repositories**: Choose whether to install it on all repositories or only specific ones.
3. **Configuration**: Once installed, Substrate automatically starts checking any Pull Requests in those repositories. It will emit a commit status check on every PR to report whether breaking changes are detected.

If a repository has OpenAPI/SQL/GraphQL spec files but no `substrate.yaml` configuration file, Substrate will prompt you to run `substrate init` to configure it.

## Enabling the Merge Blocker (Branch Protection)

Substrate emits a commit status context called `substrate/breaking-changes`. By default, GitHub shows this status, but **it does not prevent merges** until you require it.

To actually block a merge when breaking changes are detected, you must configure this status context as a **Required Status Check** in your repository's branch protection rules.

### Setup Instructions

1. **Go to Settings**: In your GitHub repository, click on the **Settings** tab.
2. **Navigate to Branches**: In the left sidebar, click on **Branches**.
3. **Edit Branch Protection Rule**: Edit the branch protection rule for your default branch (e.g., `main`). If you do not have one, click **Add branch protection rule** and set the branch name pattern to `main`.
4. **Require Status Checks**: Check the box for **Require status checks to pass before merging**.
5. **Add Substrate**: In the search bar that appears, type `substrate/breaking-changes` and select it to add it to the required list.
6. **Save Changes**: Click **Save changes** at the bottom of the page.

Once enabled, any PR with unacknowledged breaking changes (where `on_breaking_change: block`) will have its merge button disabled by GitHub.

## Troubleshooting

### Why isn't my PR blocked?

If a PR contains breaking changes but is not being blocked, check the following:

- **Check Branch Protection**: Ensure you have configured `substrate/breaking-changes` as a required status check in your branch protection rules for the target branch, as described above.
- **Check Configuration**: Ensure `on_breaking_change` is set to `block` (the default) in your `substrate.yaml` file. If it is set to `warn`, Substrate will only post an advisory comment and pass the commit status check, allowing the merge to proceed.
- **Check Setup**: Verify that the GitHub App is installed on the repository and that a `substrate.yaml` file is present (or that a setup comment was posted and the neutral commit status was set).
