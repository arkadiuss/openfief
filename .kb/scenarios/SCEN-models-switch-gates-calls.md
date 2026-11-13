---
title: With the switch or the role's toggle off the router calls no model
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:14.960Z'
id: SCEN-models-switch-gates-calls
type: scenario
---
Given the switch off, or the role's feature toggle off, when a role asks the router for an answer, then no provider is listed or called, the refusal is recorded, and the error is the switch's or the toggle's, even for a request that would be invalid or untagged.