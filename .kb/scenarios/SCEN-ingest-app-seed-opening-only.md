---
title: Only an opening batch seeds an account
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:38:56.836Z'
id: SCEN-ingest-app-seed-opening-only
type: scenario
---
Given an ingest service, when a batch runs in opening mode, then its mapping goes to the ledger's Seed and a refusal of an already seeded account fails the batch with that error; when a batch in transactions mode holds opening cash or an opening lot, then it is refused before the ledger is called.