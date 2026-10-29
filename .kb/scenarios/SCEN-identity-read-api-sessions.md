---
title: The read API accepts sessions and script tokens
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:59.602Z'
id: SCEN-identity-read-api-sessions
type: scenario
---
Given a valid session of any scope or a valid read-scoped script token, when it is presented as the bearer token to the read API, then the call is authenticated as the owner; a revoked or expired one is refused.