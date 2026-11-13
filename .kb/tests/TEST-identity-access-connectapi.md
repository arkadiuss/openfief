---
title: UpdateGrants and ListViewPlugins over Connect
status: active
sourceFile: internal/identity/connectapi/access_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
id: TEST-identity-access-connectapi
type: test
---
Serves the AuthService on a test server and checks that a demo session, no session and a caller that is no view plugin are refused UpdateGrants and ListViewPlugins, that UpdateGrants adds and ends grants in one call and answers with those in force, that a missing target is not found and an unspecified kind invalid before any change, and that ListViewPlugins shows the result with the household's ID as home.