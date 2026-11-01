---
title: A change and its event commit together
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:13.878Z'
id: SCEN-attention-outbox-same-transaction
type: scenario
---
Given the PostgreSQL store, when a change is appended and its transaction rolls back, neither the change nor its event remains; when the outbox refuses the event, the change is rolled back too; and updates and deletes of items, history and outbox are refused.