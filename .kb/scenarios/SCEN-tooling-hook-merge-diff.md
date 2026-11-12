---
title: A merge is planned from its diff, not run in full
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:38.889Z'
id: SCEN-tooling-hook-merge-diff
type: scenario
---
Given a merge being committed whose staged diff changes only tests, when the pre-commit hook runs, then the selector receives the staged changes and plans those tests' checks, and the hook does not run check:commit:full because of the merge; FULL selects it.
