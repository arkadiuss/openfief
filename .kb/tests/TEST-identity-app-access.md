---
title: Policy answers the rule for any account, and a request's context carries its caller
status: active
sourceFile: internal/identity/app/access_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T22:45:55.106Z'
id: TEST-identity-app-access
type: test
---
TestPolicyMayAct asks identityapp.Policy for every holder and action on several synthetic accounts; TestCallerInContext round-trips a caller through WithCaller and CallerFrom and finds nobody in a bare context; TestPrincipalCaller maps a passkey session, a demo session and a script token to their callers.