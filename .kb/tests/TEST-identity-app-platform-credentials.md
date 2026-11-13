---
title: Initializing the platform's credentials creates each missing Secret from random bytes and keeps each that exists
status: active
sourceFile: internal/identity/app/platform_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T07:41:44.535Z'
id: TEST-identity-app-platform-credentials
type: test
---
Runs app.InitPlatformCredentials over an in-memory SecretCreator and a deterministic random source. With both Secrets missing it creates both, the first from the first 32 bytes and the second from the next 32, as hex; with one already there it reports it kept and creates the other; with both there it creates none; a creator that fails on the second stops there, returns the first's outcome and an error that holds no value; too few random bytes create nothing.