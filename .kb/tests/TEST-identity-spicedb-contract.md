---
title: A real SpiceDB answers every holder's actions as identity.MayAct does
status: active
sourceFile: internal/identity/spicedb/contract_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:24.896Z'
id: TEST-identity-spicedb-contract
type: test
---
Under task test:authz, which starts the pinned SpiceDB in testing mode (build tag spicedb), with authz/schema.zed written and the syncer's relationships plus the household grants of today's readers: every holder kind and action on every account, and the lists, equal identity.MayAct; a sleeve grant covers an account placed later and ends when its sleeve relationship is removed or it moves, a household grant covers a later account, a viewer grant never changes, a deleted account is nobody's; reconcile is idempotent and leaves grants. Not part of task check, because building SpiceDB takes minutes. A false pass would be a schema loaded without the sleeve relation, which the sleeve rows would catch.