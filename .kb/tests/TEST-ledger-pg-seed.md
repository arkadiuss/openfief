---
title: Seeding, opening lots one at a time, interest and deletion on real PostgreSQL
status: active
sourceFile: internal/ledger/postgres/seed_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:38:23.770Z'
id: TEST-ledger-pg-seed
type: test
verification_scope: integration
---
Integration test on a disposable PostgreSQL (task test:pg): eight goroutines seed the same account with different go-live dates, so their keys differ and only the account lock can stop them; exactly one succeeds, the others get an *AlreadySeededError, and the log holds only the winner. Rerunning the winner is a no-op. LockAccount fails outside a transaction and for an unknown account. Without the lock this test fails (checked by mutation).