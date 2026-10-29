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
Under task test:pg: core enroll-link prints a one-time localhost link; the core, started in a fake cluster whose TokenReview accepts a synthetic view-plugin token for the Plugin ui (a ready view), refuses a caller with another token, enrolls a software passkey with the link's code, logs in with it, reads the read API with the session and with a script token made from it, and logs out; demo sessions are refused without the marker or the flag. Without a cluster every caller is refused: no fixed token exists (D-84). The caller check is also run with a fake TokenReview over view, connector and unknown plugins.