---
title: Only an owner session grants, revokes or lists grants, never a demo session, a script token or the read API
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:18.126Z'
id: SCEN-identity-grants-owner-only
type: scenario
---
Given a demo session, no session, an unknown session, a script token's secret and a caller that is no view plugin, when each asks to grant, revoke or list a grant, then each is refused (permission denied for a read-scoped session and for a caller that is no view plugin, unauthenticated for the others) and nothing is granted. The grant calls are on the AuthService, which only view plugins reach on the plugin port, and not on the read API. Assumes D-209, D-137 and D-211.