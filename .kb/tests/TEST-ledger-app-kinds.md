---
title: The ledger stores sleeve kinds checked by the portfolio rules and ensures the defaults on a fresh install
status: active
sourceFile: internal/ledger/app/kinds_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:52:41.964Z'
id: TEST-ledger-app-kinds
type: test
verification_scope: unit
---
Use-case tests of ledger app over the in-memory store: SaveKind adds and renames kinds and refuses a role change, EnsureDefaults stores the built-in kinds and the three default sleeves on an empty ledger, keeps the owner's renames and moves when run again, and on a ledger that already holds a sleeve of its own, such as the demo dataset's, adds the missing kinds and no default sleeve.