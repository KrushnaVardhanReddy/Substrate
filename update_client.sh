#!/bin/bash
# Modify client.go to include conclusion in CreateCheckRun
sed -i 's/CreateCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error/CreateCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary, conclusion string) error/g' api/internal/github/client.go

# Modify the implementation in client.go
sed -i 's/"conclusion": "success",/"conclusion": conclusion,/g' api/internal/github/client.go

# Update other usages
sed -i 's/err = ghClient.CreateCheckRun(/err = ghClient.CreateCheckRun(ctx, req.Org, req.Repo, req.CommitSHA, "substrate", "Substrate Trial Expired", "Substrate 90-day trial has expired. Please visit the dashboard to upgrade to the Enterprise plan and resume API protection.", "failure")/g' api/internal/webhook/push.go
# Wait, I shouldn't just replace like this. Let's do it manually with gofmt/perl or sed specifically.
