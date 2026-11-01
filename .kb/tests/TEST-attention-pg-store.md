---
title: The attention store keeps history, writes each event in the change's transaction and is append-only
status: active
sourceFile: internal/attention/postgres/store_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:38.722Z'
id: TEST-attention-pg-store
type: test
---
Runs against a fresh migrated PostgreSQL database (task test:pg). Appends a raise and an acknowledgement and reads the item and both outbox events back; marks one published; rolls back an InTx after an append and finds neither change nor event; makes the outbox refuse inserts with a test trigger and finds that the change was rolled back with it; checks that updates, deletes and truncates are refused with the append-only error and that a change out of sequence is refused; and sweeps a check over the store.