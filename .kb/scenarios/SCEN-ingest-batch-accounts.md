---
title: Completed batches keep the accounts they booked into
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T03:42:58.086Z'
id: SCEN-ingest-batch-accounts
type: scenario
---
Given a sync batch completed with accounts acc-1 and acc-2 and an upload recorded with acc-2, when the store lists completed batches, then each comes with its accounts, and a batch completed with none lists none.