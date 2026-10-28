---
title: Demo sessions need the demo marker and the flag
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:49.997Z'
id: SCEN-identity-demo-sessions
type: scenario
---
Given a core, when a view plugin asks for a demo session, then it gets a read-scoped session for the owner only when the core runs with --demo-sessions AND the database carries the demo marker; with either missing the request is refused.