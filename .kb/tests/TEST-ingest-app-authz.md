---
title: Ingest's owner services ask for the change of an account before touching it
status: active
sourceFile: internal/ingest/app/access_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T22:46:01.779Z'
id: TEST-ingest-app-authz
type: test
---
With the Owners fake denying the owner two accounts, Update, Bind, Delete, Remap onto and away from a denied account, BondLots.Add and Delete and Uploads.Upload all fail with ErrNotOwner, and no account, binding or connector call happens; the statement accounts list leaves the denied account out.