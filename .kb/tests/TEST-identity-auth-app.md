---
title: The auth use cases enroll, log in, scope, expire and revoke sessions over the real WebAuthn ceremonies
status: active
sourceFile: internal/identity/app/auth_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:36.591Z'
id: TEST-identity-auth-app
type: test
---
Runs the auth use cases over in-memory stores, the go-webauthn ceremonies adapter and a software ES256 authenticator, with a settable clock. Covers enrollment through a one-time code that logs the owner in; refusal of a missing, malformed, unknown, mismatched, expired, foreign or reused code, and a code used once across two concurrent ceremonies; passkeys working only on their host and a new host enrolled through a session-issued link; refusal of a regressed signature counter, a reused or expired ceremony and an origin off the list; 12-hour idle and 7-day absolute expiry; logout; that read sessions are forbidden every owner action; script tokens from a session; the read API accepting sessions and script tokens; and demo sessions only with both the flag and the marker.