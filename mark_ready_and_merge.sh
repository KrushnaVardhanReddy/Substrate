#!/bin/bash
export GH_TOKEN=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
for pr in 134 135 136 137; do
  gh pr ready $pr
  gh pr merge $pr --merge --delete-branch
done
