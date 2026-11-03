---
title: The built-in UI asks to confirm with a passkey and posts the refused action again
status: active
sourceFile: internal/views/ui/passkeycheck_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T02:40:59.957Z'
id: TEST-ui-passkey-check
type: test
---
Over the UI's fake core, which refuses the protected actions until a passkey check finishes: creating a script token, an enrollment link or removing a passkey shows Confirm with your passkey with the action's path and fields to post again and does nothing; /passkey-check/begin and /passkey-check/finish relay the ceremony to the core; the same post then goes ahead. A read or missing session can't begin a check; beginning a registration answers passkeyCheck true for the page's script; removing the passkey of this session drops the cookie; Sign out everywhere posts to /logout, revokes every session at the core and drops the cookie, and a read session is refused and keeps it.