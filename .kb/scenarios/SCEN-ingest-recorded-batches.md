---
title: Completed uploads and manual entries are recorded in the import batch history
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:09.104Z'
id: SCEN-ingest-recorded-batches
type: scenario
---
Given a completed upload or manual bond lot batch of a connector instance, when the core records it, then the ingest schema holds it as completed after one attempt with its source, mode, institution, counts and times, without touching the instance's sync state, and recording the same batch ID again fails.