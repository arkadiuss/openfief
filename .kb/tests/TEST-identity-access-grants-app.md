---
title: Token creation with grants, UpdateGrants and the view plugin list over in-memory fakes
status: active
sourceFile: internal/identity/app/accessgrants_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
id: TEST-identity-access-grants-app
type: test
---
Runs the auth use cases over in-memory fakes of the store, the directory and the authorization model: a token with the MCP flag and grants is made with its entries and projected, a repeated target once; an unknown or malformed target stores nothing; a model that does not answer stores nothing; a model whose write fails revokes the token with an entry and no secret; UpdateGrants adds and removes with an entry each and checks first; the view plugin list leaves out the built-in UI and is owner-only.