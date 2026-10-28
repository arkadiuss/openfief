---
title: On shutdown the core stops being ready and drains
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:32:52.562Z'
id: SCEN-core-graceful-shutdown
type: scenario
---
Given a running core, when it is told to stop, then /readyz answers 503 shutting down, the HTTP server finishes its requests, and the River workers stop after their running jobs, all within the shutdown timeout, after which what still runs is cancelled for River to retry.