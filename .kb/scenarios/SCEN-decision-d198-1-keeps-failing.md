---
title: Three failures over fifteen minutes keep failing; a success clears it
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T06:15:00.561Z'
id: SCEN-decision-d198-1-keeps-failing
type: scenario
---
Given a connector instance whose last three attempts since its last success failed, the first starting at least 15 minutes before the last ended, when its status is computed, then it keeps failing and raises one item; two failures, or three within a shorter span, don't; the next success clears it.