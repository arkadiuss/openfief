---
title: Batches with unsupported rows are listed
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:29.102Z'
id: SCEN-ingest-unsupported-rows-counted
type: scenario
---
Given an upload with skipped and unsupported rows, a clean upload and a sync with one unsupported row, the store lists the upload with two rows and the sync with one, and not the clean upload.