---
title: Seeding is recognized from the log and a re-seed is decided purely
status: active
sourceFile: internal/ledger/seed_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:38:08.342Z'
id: TEST-ledger-seed
type: test
---
Table tests of the pure ledger functions on synthetic transactions: OpeningCashID formats the opening cash key; NewOpeningCash books a positive balance as a deposit and a negative one as a withdrawal against the equity ledger account and rejects a zero balance or a key without the prefix; IsSeeding accepts opening lots and opening cash only; IsSeeded looks at current versions of the account only; DecideSeed returns SeedFirst for an unseeded account, SeedUnchanged when every incoming row equals its current version, and an *AlreadySeededError otherwise.