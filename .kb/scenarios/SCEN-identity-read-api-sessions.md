---
title: The read API authenticates a valid session or read-scoped script token as its holder and refuses a session it can't place
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:59.602Z'
id: SCEN-identity-read-api-sessions
type: scenario
---
Given a valid session of any scope (the owner's, a viewer's or the demo's) or a valid read-scoped script token, when it is presented as the bearer token to the read API, then the call is authenticated as its holder: the owner, the viewer, the demo session or the script token; a revoked or expired one is refused, and so is a session that is none of the owner's, a viewer's or the demo's, which resolves to nobody.