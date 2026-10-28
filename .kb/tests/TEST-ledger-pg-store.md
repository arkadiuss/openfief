---
title: The Postgres store migrates its own schema and round-trips every value exactly
status: active
sourceFile: internal/ledger/postgres/store_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:28:49.329Z'
id: TEST-ledger-pg-store
type: test
---
Runs against a fresh database per test on a disposable PostgreSQL (task test:pg; skipped without OPENFIEF_PG_DSN). Migrate runs again four times concurrently without error and records only 0001_ledger.sql, and the ledger schema holds exactly its six tables. Sleeves and accounts are saved, replaced and listed in ID order; an unknown account is ErrNotFound and an unknown sleeve is refused by a foreign key. One synthetic transaction of every kind, with amounts such as 123.4500, 0.000001 units, a non-ASCII description, an opening lot date and fees in another currency, plus a correction and a deposit corrected into another account, are appended and read back with Versions and Log and compared in full with go-cmp, including key order and both accounts' logs. Assertion checks are saved, one replaced, and read back per account in date and ledger account order.