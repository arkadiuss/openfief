---
title: The Access page, its token and grants forms and the owner-only rule
status: active
sourceFile: internal/views/ui/access_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
id: TEST-ui-access
type: test
---
Serves the UI over a fake core and checks the Access page lists tokens with expiry, MCP flag and grants in words and view plugins with theirs, a revoked token has no form, a core that cannot read its cluster says so, the owner-only navigation and refusals, a new token posts its name, days, MCP flag and grants once and shows its secret once, forms out of range never reach the core, the core's refusals and an unavailable model show on the page, the grants form sends exactly the changed boxes, and the Inbox item of a plugin that reads nothing links to the page.