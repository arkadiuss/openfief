---
title: A sync leaves out what an account's seeding already holds
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:13.072Z'
id: SCEN-ingest-sync-keeps-from-go-live
type: scenario
---
Given acc-1 seeded on 2026-03-02, when a sync's transactions batch delivers a deposit dated 2026-03-02, then it is not booked and the batch lists it as a skipped problem named by its key; an uploaded transactions statement with the same deposit is booked; a ledger that fails to answer fails the sync.