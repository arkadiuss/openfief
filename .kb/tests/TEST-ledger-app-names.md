---
title: The ledger's names keep a connector's first name and the owner's rename
status: active
sourceFile: internal/ledger/app/names_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T06:58:51.745Z'
id: TEST-ledger-app-names
type: test
---
app.Names over an in-memory store and a ledger whose acc-1 holds a synthetic ISIN: two deliveries keep the first name, the owner's rename wins and is cleaned; a read session, an empty name and an account the owner may not change are refused and store nothing.