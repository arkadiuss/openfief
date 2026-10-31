---
title: Sleeves and accounts are validated before they are stored
status: active
sourceFile: internal/ledger/app/ledger_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:19.114Z'
id: TEST-ledger-app-registry
type: test
verification_scope: unit
---
Table tests over the in-memory store seeded with the built-in sleeve kinds and the core sleeve. SaveSleeve stores an offensive satellite sleeve, a sleeve with no role with its kind's default role, a defensive sleeve of an offensive kind, and an offensive retirement sleeve only with the explicit offensive override; it refuses one without the override (portfolio.ErrRoleOverride), with no kind or an unstored kind (portfolio.ErrUnknownKind), and an invalid one (portfolio.ErrInvalidSleeve), and in each failing case the stored sleeves are unchanged. A new account is stored and listed in ID order, saving an account with an existing ID replaces it, an account whose sleeve is not stored fails with ErrUnknownSleeve, and an account with no institution or name fails with ledger.ErrInvalidAccount; in the failing cases the stored accounts are unchanged. Account returns the stored account by ID.