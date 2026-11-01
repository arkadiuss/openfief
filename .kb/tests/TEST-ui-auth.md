---
title: The UI signs the owner in through the AuthService and guards the browser
status: active
sourceFile: internal/views/ui/auth_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:28.085Z'
id: TEST-ui-auth
type: test
verification_scope: integration
---
Runs the UI handler against an in-memory AuthService and read API over Connect, with the SDK's CoreHTTPClient over a token file the test rotates. Asserts the login page offers View demo only when the core says so, the passkey and enrollment ceremonies relay the core's WebAuthn JSON untouched, the session cookie is __Host-, HttpOnly, Secure and SameSite=Strict, logout revokes and drops it, every AuthService call carries the current plugin token, cross-origin state-changing requests are refused, and every page carries the strict CSP with no inline script or style. A false pass would need the fake to accept any bearer, which it doesn't.