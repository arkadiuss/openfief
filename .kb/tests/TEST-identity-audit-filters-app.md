---
title: The audit log filtered by outcome and by a set of actions, and viewer sessions named by person, over in-memory stores
status: active
origin:
  kind: agent
  ref: 'S7 task brief: the Access page''s history and Settings'' audit log view (D-210, D-211), 2026-11-13'
  recorded_at: '2026-11-14T00:30:00.000Z'
sourceFile: internal/identity/app/auditview_test.go
verification_scope: integration
id: TEST-identity-audit-filters-app
type: test
---
Runs the identity use cases over in-memory stores. Checks the entries of the history actions newest first, the refused outcome alone, an unknown outcome refused, and the person of a viewer's session on a page.