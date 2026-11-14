---
title: The Audit log page has words for every action, actor and record of the contract and reads changes as IDs
status: active
sourceFile: internal/views/ui/audit_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:44.141Z'
id: TEST-ui-audit-names-changes
type: test
---
Table tests over the contract's enums: every action, kind of actor and kind of record has words, and an owner's change and the operator's install read as who, what and records by name or ID.