---
title: Removing a passkey revokes the sessions it signed in
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:09.395Z'
id: SCEN-identity-remove-passkey-revokes-sessions
type: scenario
---
Given two passkeys and a session signed in by each, when the owner removes the first passkey, then every session it signed in is revoked in the same transaction and the other passkey's session goes on; when the removed passkey signed in the caller's own session, the core says so and the UI drops the cookie and opens the login page.