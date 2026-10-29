---
title: A failed sync is retried with the same batch after a doubling backoff
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:12.545Z'
id: SCEN-ingest-retry-same-batch
type: scenario
---
Given a sync attempt that fails, when the next attempts run after the first retry delay, doubled for each further failure up to the maximum, then each calls the connector with the same batch ID, mode and watermark; when one succeeds, then the batch completes, the watermark is stored with it and the failures reset; the watermark is never stored for a failed attempt.