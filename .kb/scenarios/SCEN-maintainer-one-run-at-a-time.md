---
title: A second run is not started while one is queued or running
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:30.230Z'
id: SCEN-maintainer-one-run-at-a-time
type: scenario
---
Given a run queued or running, when the owner asks again or the schedule fires, then nothing new is queued; a run that stayed active for half an hour is closed as failed so a lost job never blocks the maintainer.