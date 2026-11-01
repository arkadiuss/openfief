---
title: The hook runs the selector's plan
status: active
sourceFile: scripts/tests/check-commit.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-30T03:32:31.735Z'
id: TEST-tooling-check-commit
type: test
verification_scope: integration
---
Runs scripts/check-commit.sh in a throwaway git repository with a fake go and a fake task: the plan's steps run in order with a variable holding spaces kept as one argument, and a task reading standard input cannot swallow the next line; the tool receives the staged changes and both modules' package listings; a failing step stops the rest and fails the hook; a failing selector runs nothing; FULL and a merge being committed select the full check; DRY_RUN plans without running.