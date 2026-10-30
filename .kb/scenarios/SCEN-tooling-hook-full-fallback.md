---
title: When in doubt, the hook runs everything
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T03:32:54.558Z'
id: SCEN-tooling-hook-full-fallback
type: scenario
---
Given a staged change to go.mod, the Taskfile, the lint configuration, a tool or script, a contract, generated code, embedded design assets, a file no rule names or no package reads, or a merge being committed, when the pre-commit hook runs, then it runs the full check:commit:full and says why.