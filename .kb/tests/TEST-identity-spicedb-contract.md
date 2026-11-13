---
title: A real SpiceDB answers every holder's actions as identity.MayAct does, and a grant projected through the adapters covers what D-209 says and only lets its holder view
status: active
sourceFile: internal/identity/spicedb/contract_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:24.896Z'
id: TEST-identity-spicedb-contract
type: test
---
Under task test:authz, against a real SpiceDB in testing mode with authz/schema.zed written. With the household grants the cutover writes for the holders of read access, every holder's every action on every account is answered as identity.MayAct does, and the lists agree. A sleeve grant covers an account placed in the sleeve later and stops when the account's sleeve relationship is removed or the account moves; a household grant covers an account added later; an account grant covers that account only; no grant ever lets its holder change. Revoking a grant, by removing its relationship, ends what only it gave. Reconcile writes the grants of the table and removes one the table does not hold, and a second Reconcile writes nothing.