---
title: The owner logs in with a passkey; a cloned one is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:51.751Z'
id: SCEN-identity-passkey-login
type: scenario
---
Given a passkey created on a host, when the owner signs the login challenge with it on that host, then the core verifies the assertion with go-webauthn, records the counter and issues an owner session; when an assertion carries a counter that did not go up, the login is refused as a possible clone.