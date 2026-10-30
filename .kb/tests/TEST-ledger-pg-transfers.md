---
title: PostgreSQL stores transfer links append-only and serializes decisions
status: active
sourceFile: internal/ledger/postgres/transfers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T06:05:33.687Z'
id: TEST-ledger-pg-transfers
type: test
---
On a fresh migrated database (task test:pg): links round-trip in pair then version order, a stored version is a version conflict, a version without its predecessor fails with OFA02, UPDATE, DELETE and TRUNCATE of ledger.transfer_links fail with OFA01, LockTransferLinks needs InTx, and two concurrent confirmations of one deposit through app.Ledger store exactly one link, which a rejection then supersedes as version 2.