---
title: CLI Reference
---

## 3. CLI Reference

If you are running Substrate locally or in a custom pipeline, use the CLI:

### `substrate diff`
Compares two schema files and exits with code `2` if a breaking change is found.
```bash
substrate diff --base=main.yaml --head=pr.yaml --type=openapi
```

### `substrate serve`
Starts the Diff Engine in HTTP Server mode (used by the GitHub App worker).
```bash
substrate serve --port=8080
```

### `substrate init`
Scaffolds a `substrate.yaml` configuration file interactively.
```bash
substrate init
```

---
