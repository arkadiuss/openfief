---
title: A passkey login is relayed to the core and ends in a session cookie
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:48.123Z'
id: SCEN-ui-passkey-login
type: scenario
---
Given the core's BeginLogin returns a ceremony and request options, when the browser posts to /login/passkey/begin and then posts the credential's toJSON() to /login/passkey/finish, then the UI returns the options untouched, hands the credential JSON to FinishLogin, and stores the new session in the cookie; a refused assertion sets no cookie.