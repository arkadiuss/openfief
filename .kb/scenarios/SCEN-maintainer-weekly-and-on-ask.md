---
title: The schedule runs every Monday and only the owner may ask
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:29.374Z'
id: SCEN-maintainer-weekly-and-on-ask
type: scenario
---
Given the maintainer on, when it is Monday 06:30 in the core's zone then the weekly job queues and works a run of its own; when the owner asks a run is queued at once and its job works it; a read session or a viewer asking is refused before anything is queued.