---
title: Uploads choose the mode from the ledger and list the accounts whose connector takes files
status: active
sourceFile: internal/ingest/app/fileaccounts_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T04:10:39.512Z'
id: TEST-ingest-app-file-accounts
type: test
verification_scope: unit
---
With in-memory fakes, uploads for an unseeded and a seeded account are handed to the connector in opening and transactions mode, the owner's override wins, and an unreadable seeding imports nothing; FileAccounts lists only accounts bound to exactly one running instance whose manifest declares statement files, each with its next mode, refuses a reader, and lists none before any instance is found.