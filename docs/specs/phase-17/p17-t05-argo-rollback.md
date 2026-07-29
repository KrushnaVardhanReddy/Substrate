# P17-T05: Auto-Rollback via ArgoCD/Flux

## Overview
Wire Substrate's `Can-Rollback` and eBPF Drift engine to ArgoCD/Flux webhooks, automatically reverting a deployment if it causes severe schema violations in production.

## Goals
1. Provide an endpoint `POST /api/v1/argo/drift-webhook` that ArgoCD/Flux can hit post-deployment.
2. Evaluate the drift anomalies (`drift_anomalies` table) for the deployed service.
3. If severe schema violations are detected, return a 406 or trigger a rollback API call to ArgoCD.

## Requirements
- Go Backend: `api/internal/handlers/argocd.go` to handle webhooks.
- Create tests mocking ArgoCD API.
