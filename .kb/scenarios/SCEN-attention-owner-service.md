---
title: Only the owner's view reaches the AttentionService
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:21.203Z'
id: SCEN-attention-owner-service
type: scenario
---
Given the core's plugin port, when the owner's view acknowledges an unknown item the core answers not found, while a caller that is not a view plugin and an unknown session are refused before the items are read.