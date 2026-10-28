---
title: A read-scoped session can only read
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:48.178Z'
id: SCEN-identity-read-scope
type: scenario
---
Given a read-scoped session, when it calls the read API, WhoAmI or Logout, then the call succeeds; when it tries to create a script token or an enrollment link, list or add passkeys, then the core refuses with permission denied.