#!/bin/bash
export GH_TOKEN=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
git config --global user.email "bot@antigravity.ai"
git config --global user.name "Antigravity Bot"

for pr in 134 135 136 137; do
  echo "Processing PR $pr..."
  
  # Ensure clean state
  git reset --hard
  git checkout feature/dev
  git pull origin feature/dev

  gh pr checkout $pr || continue
  git fetch origin feature/dev
  
  if ! git merge origin/feature/dev -X theirs -m "Auto-merge origin/feature/dev into PR $pr"; then
     echo "Conflict detected! Resolving..."
     # For tasks.md, keep feature/dev
     git checkout --ours tasks.md Makefile dashboard/tests/dashboard.spec.ts 2>/dev/null || true
     git add tasks.md Makefile dashboard/tests/dashboard.spec.ts 2>/dev/null || true
     
     # For everything else, keep the PR's changes
     git checkout --theirs . 2>/dev/null || true
     git add .
     git commit -m "Auto-resolve conflicts favoring PR changes"
  fi
  
  git push origin HEAD
  gh pr merge $pr --merge --delete-branch
done
