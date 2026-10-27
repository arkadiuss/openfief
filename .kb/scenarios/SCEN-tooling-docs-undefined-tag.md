---
title: An undefined tag fails the docs check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:05.659Z'
id: SCEN-tooling-docs-undefined-tag
type: scenario
---
Given docs/DESIGN.md references OQ-999, which no item defines, when task docs:check runs, then it reports the undefined OQ tag and exits non-zero.
