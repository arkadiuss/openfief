---
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T02:54:48.810Z'
title: The hook step clears dead runs stores and runs both staged checks
sourceFile: scripts/tests/kibi-staged.test.sh
verification_scope: integration
id: TEST-tooling-kibi-staged
type: test
---
Runs scripts/kibi-staged.sh with a fake kibi that logs its calls, in a temporary directory holding a store of a dead process, one of a live process and unrelated directories: only the dead process's store is removed; check-generated and check --staged run in that order; a failing check-generated fails the step and the staged check does not run.