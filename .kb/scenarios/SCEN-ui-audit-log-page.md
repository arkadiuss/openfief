---
title: The owner reads and filters the audit log in Settings
status: active
origin:
  kind: agent
  ref: 'S7 task brief: the Access page''s history and Settings'' audit log view (D-210, D-211), 2026-11-13'
  recorded_at: '2026-11-14T00:30:00.000Z'
id: SCEN-ui-audit-log-page
type: scenario
---
Given the owner's session, Settings' Audit log lists every entry newest first with when, who, what, which records by name (by ID when gone) and the outcome, 50 to a page; the filters by actor kind, action, outcome and days narrow it and are kept when paging; a malformed filter is refused before the core; a viewer, a read session and the demo's session get 403 and no navigation link; an unavailable core is said so.