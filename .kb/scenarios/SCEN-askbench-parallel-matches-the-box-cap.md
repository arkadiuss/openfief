---
title: A bench run without -parallel asks three questions at once, and the flag still sets it
status: active
origin:
  kind: agent
  ref: 'Design session decision: bench parallel 3 (D-208) and the performance tool''s description (D-203), 2026-11-15'
  recorded_at: '2026-11-15T10:08:39.071Z'
id: SCEN-askbench-parallel-matches-the-box-cap
type: scenario
---
Given no -parallel flag, when the bench reads its arguments, then three questions are asked at once; given -parallel 1, then one.
