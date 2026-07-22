# Phase 10 - Task 16: Webhook Auto-Discovery Integration

## 1. Goal
Wire the Phase 5 Discovery Engine into the Cloudflare Webhook pipeline to automatically scan repos for dependencies, eliminating manual `substrate.yaml` config.

## 2. Requirements
- The webhook worker should fetch `package.json`, `go.mod`, etc.
- Upsert implicit dependencies directly into the DB.
