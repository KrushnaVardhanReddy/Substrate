---
title: Contract Registry
description: Contract Registry & Cross-Repo Impact Analysis
---

The Substrate Contract Registry (Phase 3) enables **Cross-Repo Impact Analysis**.

1. When a consumer (e.g., `iOS-App`) pushes its `substrate.yaml` to `main`, Substrate records its dependency on your API.
2. When you open a PR on the API, Substrate checks the Registry.
3. If your change breaks the `iOS-App`, the GitHub App will print a **Compatibility Matrix** directly in your PR comment, explicitly listing `iOS-App` as a blocked consumer.