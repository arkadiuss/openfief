---
title: Over Connect, D-183's refusal carries the PasskeyCheckRequired detail
status: active
sourceFile: internal/identity/connectapi/passkeycheck_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T02:40:57.565Z'
id: TEST-identity-passkey-check-connectapi
type: test
---
Serves the AuthService over an httptest server with a stale owner session planted in the store: CreateScriptToken, CreateEnrollmentLink, RemovePasskey and BeginRegistration with the session fail with FAILED_PRECONDITION and the PasskeyCheckRequired error detail naming the five-minute window; BeginPasskeyCheck returns options; an unknown check ceremony fails as a failed precondition; a read session can't begin a check or sign out everywhere; SignOutEverywhere revokes both sessions and WhoAmI then fails as unauthenticated.