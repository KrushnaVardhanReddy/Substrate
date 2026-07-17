# Phase 10 - Task 10: Consumer-Driven Contract Manifests

## 1. Goal
Allow frontend apps to upload `.substrate-consumer.yaml` declaring required fields, competing with PactFlow.

## 2. Requirements
- Let consumers declare which fields of an OpenAPI spec they *actually* use.
- Ignore breaking changes if no consumer uses the dropped field.
