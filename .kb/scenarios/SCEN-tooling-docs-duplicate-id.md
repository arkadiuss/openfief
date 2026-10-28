---
title: An ID defined twice fails the docs check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T03:53:25.108Z'
id: SCEN-tooling-docs-duplicate-id
type: scenario
---
Given OQ-21 is defined in docs/DESIGN.md and again in a proposal file, when task docs:check runs, then it reports OQ-21 as defined more than once and exits non-zero.
