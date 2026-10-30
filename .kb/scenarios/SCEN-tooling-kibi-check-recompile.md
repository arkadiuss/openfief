---
title: A compile defect is recompiled once, an authored violation never
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T04:09:46.727Z'
id: SCEN-tooling-kibi-check-recompile
type: scenario
---
Given a compiled store that lacks a relationship the authored .kb files hold, when task kibi:check runs, then the store is compiled again from the same files and checked once more, and that result is final; given any other violation, the gate fails at once without a recompile. Assumes the authored files do not change between the two checks.