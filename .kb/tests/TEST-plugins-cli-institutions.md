---
title: The CLI warns about an institution another instance books for
status: active
sourceFile: internal/plugins/cli/institutions_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T07:22:17.231Z'
id: TEST-plugins-cli-institutions
type: test
---
Renders an install summary and an upgrade summary for a connector booking for ibkr while another instance books for it: both carry the warning naming that instance and institution, the install lists the institutions, the upgrade lists the added institution in its diff, and a summary with no shared institution warns about nothing.