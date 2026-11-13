---
title: The grants table keeps one grant in force per holder and target and the history after, refuses unfit IDs, and shares a transaction with the audit entry
status: active
sourceFile: internal/identity/postgres/grants_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:09.363Z'
id: TEST-identity-grants-postgres
type: test
---
Under task test:pg, against a real PostgreSQL. The store saves a grant once while it is in force and again after it ended, lists by holder and by target oldest first, ends a holder's or a target's grants and returns them, never ending one before its creation. The table's own checks refuse an unknown kind, a token named by a sentence or in capitals, a plugin in capitals or too long, an account with a space, a household that is not home, a grant revoked before it was made and a second grant in force. The audit log's action list takes grant_created and grant_revoked and refuses another. The grant use cases over the real store leave the entries a revoked token's grants and each grant should, and with an audit log that refuses entries no grant is stored; the cutover is made once.