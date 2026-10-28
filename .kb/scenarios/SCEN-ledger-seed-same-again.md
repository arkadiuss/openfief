---
title: Re-running the same seeding is a no-op
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:38:46.022Z'
id: SCEN-ledger-seed-same-again
type: scenario
---
Given an already seeded account, when the identical seeding is run again, so that every incoming transaction of the account equals its current version, then nothing is appended and the import summary lists the account as already seeded instead of failing.