---
title: The Ask box's calls at once are bounded
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:37:58.805Z'
id: SCEN-models-ask-concurrent-calls
type: scenario
---
Given two places and a short wait, when two calls hold them then a third is refused as busy with an audit entry of class busy, and with a long wait it takes the place a finished call frees; a failed call frees its place and a caller that leaves while waiting is not audited.