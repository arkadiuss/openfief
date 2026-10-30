---
title: A docs-only commit skips Go
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T03:32:50.669Z'
id: SCEN-tooling-hook-docs-only
type: scenario
---
Given a staged change to pages under docs/ or records under .kb/ only, when the pre-commit hook runs, then no Go format, lint, vulnerability or test step runs, while docs:check, docs:build and kibi's staged check run and the commit-msg hook still runs the docs staleness gate.