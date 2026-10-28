---
title: The ui plugin's manifest asks for the view capability only
status: active
sourceFile: plugins/ui/main_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:38.330Z'
id: TEST-ui-plugin
type: test
---
Parses plugins/ui/manifest.json with the operator's own label parser and compares it with a manifest that has only the view capability, no egress host and no Secret, and checks that a config schema exists. Checks that the read API's HTTP client never uses a proxy and has a timeout, since plugins reach the core directly.