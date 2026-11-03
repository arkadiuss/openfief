---
title: Sessions record their passkey and last passkey check, with D-183's five-minute window
status: active
sourceFile: internal/identity/passkeycheck_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T02:40:53.009Z'
id: TEST-identity-passkey-check-domain
type: test
---
Table tests of Session.CheckRecentPasskey at the edges of the five-minute window (just within, exactly five minutes, long after, no check), and of SignedInWithPasskey, PasskeyChecked and SignedInBy: signing in with a passkey records a copy of its credential ID and counts as a check at the session's creation, a check never moves back, and a demo session can neither record a passkey nor be checked. Pure functions over synthetic sessions; a false pass would need the window compared with the wrong clock.