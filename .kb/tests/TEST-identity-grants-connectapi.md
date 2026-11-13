---
title: CreateGrant, RevokeGrant and ListGrants over Connect are the owner's alone and map the grant kinds and errors
status: active
sourceFile: internal/identity/connectapi/grants_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:10.054Z'
id: TEST-identity-grants-connectapi
type: test
---
Serves the AuthService over an httptest server with a fake store, directory and an owner session, and calls it with a generated client as a view plugin. A demo session, no session, an unknown session and a caller that is no view plugin are refused every grant call and nothing is granted. The owner grants an account, a sleeve and the household, gets the grant in force back when granting again, and the codes of what can't be granted: not found for an unknown account, sleeve or plugin, invalid argument for a secret as a target, a person, a missing or unknown kind and a household that is away. Revoking changes nothing the second time and a grant never made is not found; the list filters by holder and by whether the grant is in force and keeps the history; the audit log shows the grants by kind and ID.