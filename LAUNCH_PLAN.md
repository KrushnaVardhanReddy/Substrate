# Post-T04 Execution Plan: Launching the MVP

This document outlines the exact sequence of steps to take immediately after the Phase 1a T04 (End-to-End Tests) PR is merged. 

## Step 1: Package as a GitHub Action (P1-MVP)
We need to convert our raw Go binary into a usable GitHub Action.
1. Create a `Dockerfile` that builds the Go binary.
2. Create an `action.yml` file in the root directory that defines the inputs (e.g., `base_schema`, `head_schema`, `config_file`).
3. Create a simple `entrypoint.sh` script that executes `substrate diff` inside the Docker container.

## Step 2: Dogfooding (Testing in CI)
Before launching to the public, we must prove the action works.
1. Create a `.github/workflows/test-action.yml` workflow in this repository.
2. Configure it to run the local `action.yml` against our `testdata/rev_breaking.yaml` files.
3. Verify that the GitHub Action correctly fails the build with a `❌ BREAKING CHANGE` message.

## Step 3: Documentation & Branding
1. Update `README.md` to include installation instructions for the GitHub Action.
2. Create an attractive branding logo for the GitHub Marketplace listing.

## Step 4: Publish to Marketplace
1. Draft a new Release on GitHub.
2. Check the box to "Publish this Action to the GitHub Marketplace".
3. The MVP is officially live!

## Step 5: Distribution
1. Write the "Optic Alternative" blog post/SEO article.
2. Share the repository on HackerNews ("Show HN: We built a lightning-fast OpenAPI diff engine in Go").
