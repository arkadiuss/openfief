---
title: With the switch, the toggle or the model portfolio off nothing runs
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:28.564Z'
id: SCEN-maintainer-off-runs-nothing
type: scenario
---
Given the switch off, or the maintainer's toggle off, or the model portfolio off, when the owner asks or the schedule fires, then no run is queued, no figure is read, no model is called and no audit entry is written; a run queued before the switch went off ends refused without a call.