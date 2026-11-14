---
title: The Access page's history and Settings' audit log over a fake core
status: active
origin:
  kind: agent
  ref: 'S7 task brief: the Access page''s history and Settings'' audit log view (D-210, D-211), 2026-11-13'
  recorded_at: '2026-11-14T00:30:00.000Z'
sourceFile: internal/views/ui/audit_test.go
verification_scope: integration
id: TEST-ui-audit
type: test
---
Serves the UI over fake Connect servers with a twelve-entry audit log. Checks the history's rows, order and names (a record that is gone by its ID), its paging and its request, the Audit log page's rows, the filters by actor kind, action, outcome and days reaching the core as set, malformed filters never reaching it, paging with the filter kept, the unavailable core, and that viewers, read and demo sessions get 403 and no link.