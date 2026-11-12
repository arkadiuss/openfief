---
title: Only an owner session lists the audit log, filtered by actor, action and time
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T19:17:43.222Z'
id: SCEN-identity-audit-owner-view
type: scenario
---
Given the AuthService on the plugin port, when a view plugin calls ListAuditEntries with an owner session it gets a newest-first page filtered by actor, action and time range; a demo session, a script token or no session is refused, and the log is not on the read API. The Settings page that shows it is not built, so the requirement stays open.