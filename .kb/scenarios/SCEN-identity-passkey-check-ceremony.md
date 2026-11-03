---
title: A passkey check is an assertion in the signed-in session
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:04.778Z'
id: SCEN-identity-passkey-check-ceremony
type: scenario
---
Given an owner session on a host, when the owner begins a passkey check and the browser answers with an assertion from a passkey of that host, then the core verifies it as a login's, records the passkey's use and the check's time on the session, and the protected actions are allowed for five more minutes; a login ceremony can't finish a check, a check's ceremony can't log in, an assertion for another host fails and records nothing, and a read session can't begin one. Signing in with a passkey, or creating one from an enrollment code, counts as the session's first check.