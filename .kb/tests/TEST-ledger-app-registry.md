---
title: Sleeves and accounts are validated before they are stored
status: active
sourceFile: internal/ledger/app/ledger_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:19.114Z'
id: TEST-ledger-app-registry
type: test
---
Table tests over the in-memory store: a valid offensive sleeve is stored next to the seeded core sleeve and one with no role fails with portfolio.ErrInvalidSleeve and is not stored. A new account is stored and listed in ID order, saving an account with an existing ID replaces it, an account whose sleeve is not stored fails with ErrUnknownSleeve, and an account with no institution or name fails with ledger.ErrInvalidAccount; in the failing cases the stored accounts are unchanged. Account returns the stored account by ID.