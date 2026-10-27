---
title: An edited ground rule fails the docs check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:20.161Z'
id: SCEN-tooling-docs-ground-rule-edited
type: scenario
---
Given the text of GR-14 in docs/DESIGN.md differs from HEAD, when task docs:check runs, then it reports the changed ground-rule line and exits non-zero.
