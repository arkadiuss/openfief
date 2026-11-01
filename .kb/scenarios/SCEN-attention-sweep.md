---
title: The core sweeps every check and deduplicates
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:12.011Z'
id: SCEN-attention-sweep
type: scenario
---
Given two checks of which one fails, when the core sweeps twice, the other check's findings raise one item each with one event, and the failing check's error is reported without stopping the rest; a single condition raised and cleared through the push port follows the same rules.