---
title: An import batch records where its data came from and what became of it
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:17:31.524Z'
id: SCEN-ingest-batch-record
type: scenario
---
Given a batch started for institution ibkr from source upload in transactions mode on 2026-03-02, when it completes with a mapping of three transactions, two balance assertions and one unsupported row, and an import summary of two new transactions, one unchanged and one failed assertion, then the batch holds those counts, that summary and the problem, independent of later changes to the mapping; a batch without an ID, institution, source, mode or start date cannot be started.