---
title: An account started from its whole history is never seeded
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:32.375Z'
id: SCEN-ingest-start-from-whole-history
type: scenario
---
Given an account that holds transactions from a statement of its whole history but isn't seeded, when the owner uploads another statement, then the core chooses transactions mode, lists the account with that mode, and refuses an opening upload for it, so its holdings are never counted twice.