---
title: View demo starts a read session and logout ends it
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:52.690Z'
id: SCEN-ui-demo-and-logout
type: scenario
---
Given the core offers demo sessions, when the browser posts to /login/demo, then the read-scoped session goes in the cookie and the overview opens; when the core refuses, no cookie is set. When the owner posts to /logout, the core revokes the session and the cookie is deleted.