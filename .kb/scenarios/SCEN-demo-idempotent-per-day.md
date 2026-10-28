---
title: Loading the same day again changes nothing; another day needs a reset
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:48.581Z'
id: SCEN-demo-idempotent-per-day
type: scenario
---
Given a demo database loaded for 2026-10-28, when the loader runs again for that day, then it succeeds and the number of transaction versions is unchanged; when it runs for 2026-10-29 without --reset-demo, then it fails asking for --reset-demo, because the ledger refuses to re-seed an account; with --reset-demo it loads the new day.