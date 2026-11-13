---
title: On kind and k3d, with the core on SpiceDB, a token and a viewer granted one sleeve read it alone and a revoked viewer's session stops
status: active
origin:
  kind: agent
  ref: 'S6b task brief: the Access page''s People section and the viewers end-to-end test (D-65, D-183, D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T22:00:00.000Z'
sourceFile: test/e2e/viewers_test.go
verification_scope: end_to_end
id: TEST-e2e-viewers
type: test
---
Runs against the real stack. The owner's session is a virtual authenticator's passkey. CreateScriptToken with one sleeve grant lists that sleeve's account alone; InviteViewer's link enrolls a viewer's passkey through the AuthService, which is spent afterwards, and the viewer signs in with it, lists that account alone through the read API and the UI's accounts page, is refused CreateAccount and the Access page, and after RevokePerson their session no longer reads.