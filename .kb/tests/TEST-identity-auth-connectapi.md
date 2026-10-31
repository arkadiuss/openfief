---
title: The AuthService handler admits only verified callers and maps errors to Connect codes
status: active
sourceFile: internal/identity/connectapi/handler_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:42.344Z'
id: TEST-identity-auth-connectapi
type: test
verification_scope: integration
---
Serves the AuthService over httptest with a fake caller verifier and a demo database. A call without a caller token or with a refused one is permission_denied, a verifier failure is internal; a demo session is read-scoped, may call WhoAmI and Logout, and gets permission_denied for script tokens, passkeys and enrollment links; an origin off the list is permission_denied, an unknown ceremony failed_precondition, and a revoked or missing session unauthenticated.