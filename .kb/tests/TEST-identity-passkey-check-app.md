---
title: The auth use cases refuse D-183's actions without a recent passkey check
status: active
sourceFile: internal/identity/app/passkeycheck_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T02:40:55.266Z'
id: TEST-identity-passkey-check-app
type: test
---
Over in-memory stores, the real go-webauthn ceremonies and a software authenticator with a settable clock: within five minutes of an enrollment or a login, creating a script token, issuing an enrollment link and beginning a registration with a session succeed; from five minutes on each fails with ErrPasskeyCheckRequired while listing passkeys and tokens still works; a passkey check (BeginPasskeyCheck, FinishPasskeyCheck) opens five more minutes; removing a passkey needs one too; a demo session can't begin a check; a login ceremony can't finish a check nor the other way round; an assertion for another host fails and records nothing; signing out everywhere revokes every session on every host, the caller's own included, needs no check and refuses a read session.