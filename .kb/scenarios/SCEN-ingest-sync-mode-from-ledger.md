---
title: A newly mapped account is seeded while the seeded ones sync in transactions mode
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:10.033Z'
id: SCEN-ingest-sync-mode-from-ledger
type: scenario
---
Given an instance bound to acc-1, seeded, and acc-2, newly bound and not seeded, with watermark w-0, when it is due, then the core runs an opening batch and then a transactions batch, both asking from w-0; once acc-2 is seeded the next sync is a transactions batch alone from the stored watermark; an opening batch that completed before its transactions batch failed is followed by a retry of the transactions batch with the same ID.