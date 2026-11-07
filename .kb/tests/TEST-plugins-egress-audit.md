---
title: AuditLog passes decision lines to granted hosts and reports every broken rule without quoting a value
status: active
sourceFile: internal/plugins/smokescreen/audit_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T03:30:39.510Z'
id: TEST-plugins-egress-audit
type: test
---
Table tests feed AuditLog synthetic Smokescreen logs with decision lines and other lines: allowed CONNECTs to a granted host and refusals pass and are counted by instance and host; an allowed request to a host or port off the instance's grant, or by an instance with no grant, is a problem; a line with an extra field such as a proxy authorization, a requested host holding a path and query, or no role or allow field is a problem whose text names the field but never its value. TestDecisionLines checks that only decision lines are read.