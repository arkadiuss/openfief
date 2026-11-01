---
title: The core's owner-session port asks the authorizer about the session's owner for each account
status: active
sourceFile: cmd/core/owners_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T22:46:05.808Z'
id: TEST-core-owner-sessions
type: test
---
TestOwnerSessionsChangeable runs ownerSessions.Changeable over a fake session check and a recording authorizer: all accounts, one denied, no authorizer (none), an authorizer that can't tell, a read session, no session and an origin off the list; the authorizer is asked only about the owner of the session.