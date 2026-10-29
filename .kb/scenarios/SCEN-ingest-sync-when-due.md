---
title: Each running instance syncs when its sync policy says it is due
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:08.500Z'
id: SCEN-ingest-sync-when-due
type: scenario
---
Given an instance that never synced, when the scheduler ticks, then it syncs at once in the batches its bound accounts' seeding calls for; a tick six hours later, before the interval plus its ten minutes of jitter, starts nothing; a tick after it syncs in transactions mode from the stored watermark once the account is seeded, and the next sync is due after the interval plus jitter again. The policy refuses a non-positive interval or first retry, a negative jitter and a maximum retry below the first.