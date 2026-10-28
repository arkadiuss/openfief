---
title: Sleeve kinds are checked before they are stored, and the defaults are ensured once
status: active
sourceFile: internal/ledger/app/kinds_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:52:41.964Z'
id: TEST-ledger-app-kinds
type: test
---
Table tests over the in-memory store seeded with the built-in kinds. SaveKind stores a custom kind in key order, renames a built-in kind, and refuses changing core's default role and an invalid key with the portfolio errors, leaving the stored kinds unchanged, which Kinds returns in full. EnsureDefaults on an empty store, run twice, stores exactly the six built-in kinds and the sleeves Bonds, Core and Satellite; after the owner renames the core kind and moves the satellite sleeve to agent-pool, running it again keeps both changes. A store that duplicated or overwrote defaults would fail the full go-cmp comparison.