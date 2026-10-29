---
title: The core serves the AuthService end to end and admits only view plugins as callers
status: active
sourceFile: cmd/core/auth_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:44.211Z'
id: TEST-core-auth
type: test
---
Under task test:pg on a fresh database: core enroll-link prints a one-time localhost link; the core, started as serve starts it with a dev caller token, refuses a caller with another token, enrolls a software passkey with the link's code, logs in with it, serves the read API with the session and with a script token created from it, and refuses the session after logout. Demo sessions are refused on a database without the demo marker and without --demo-sessions, and a demo session reads but can't create script tokens or enrollment links. Without a database, the cluster caller check over a fake TokenReview accepts only a view plugin's token and refuses a connector, the core itself, an unknown token and a plugin without a resource.