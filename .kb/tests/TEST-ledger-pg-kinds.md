---
title: PostgreSQL stores sleeve kinds, backfills old sleeves and refuses role changes
status: active
sourceFile: internal/ledger/postgres/kinds_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:52:47.461Z'
id: TEST-ledger-pg-kinds
type: test
---
Runs on a disposable PostgreSQL through task test:pg. After migrating, Kinds equals portfolio.DefaultKinds. A database migrated only through 0001 gets a defensive and an offensive sleeve, and after the full migration they read back with kinds core and satellite. SaveKind stores a custom kind and a rename, keeps bonds defensive when asked to make it offensive, and a direct UPDATE of a default role fails with SQLSTATE 23000 from the trigger. Saving a sleeve with an unstored kind fails the foreign key (23503) and an offensive override on a defensive sleeve fails the check (23514). Four concurrent EnsureDefaults calls succeed, and a later run keeps the owner's change to the core sleeve.