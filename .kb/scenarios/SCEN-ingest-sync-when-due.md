---
title: Each running instance syncs when its sync policy says it is due
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:08.500Z'
id: SCEN-ingest-sync-when-due
type: scenario
---
Given a six-hour interval with jitter, when an instance has never synced, then the next tick syncs it in opening mode; when it has synced, then it is not synced again before the interval plus jitter, and then it syncs in transactions mode with the watermark the last sync returned.