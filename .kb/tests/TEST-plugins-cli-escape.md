---
title: The install and upgrade screens print vendor text with visible escapes
status: active
sourceFile: internal/plugins/cli/escape_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T09:18:02.910Z'
id: TEST-plugins-cli-escape
type: test
---
TestRenderEscapesVendorText renders an install and an upgrade of a manifest carrying the review's payload in a schema property name, a description, a Secret name and a mount path, and an override in an enum value, as if it had passed validation: neither screen may hold a raw control or bidi character, and each shows the payload Go-quoted, the override as \u202e and the mount path's escape as \x1b.