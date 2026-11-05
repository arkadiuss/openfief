---
title: The core runs one connector sync at a time unless the owner raises the worker count
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T18:23:46.417Z'
id: SCEN-core-one-sync-worker
type: scenario
---
Given three connector instances due at once, when the scheduler ticks with no worker count set or with one, then it starts one sync and the others sync on later ticks; with two workers it starts two. The core's serve command defaults --sync-workers to 1 and refuses a value below one.