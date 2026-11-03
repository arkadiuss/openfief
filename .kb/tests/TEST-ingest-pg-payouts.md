---
title: PostgreSQL keeps each account's payout setting with its history in the ingest schema and forgets it with the account
status: active
sourceFile: internal/ingest/postgres/payouts_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T11:49:04.685Z'
id: TEST-ingest-pg-payouts
type: test
---
Runs against a fresh database with the ingest migrations under task test:pg. An account with no stored setting reads the default, as cash; a paid-out setting with an earlier cash one, a change back to cash from a later day and a first paid-out setting each round-trip exactly, replacing the stored one; and ForgetAccount removes it, so the default reads again.