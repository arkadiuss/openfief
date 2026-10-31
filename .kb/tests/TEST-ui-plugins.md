---
title: The UI shows installed plugins and their trust to both scopes, read-only, and marks removed instances
status: active
sourceFile: internal/views/ui/plugins_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T04:02:31.828Z'
id: TEST-ui-plugins
type: test
verification_scope: integration
---
Handler tests of the UI against in-memory Connect fakes: the owner's settings page and the /plugins page list each instance with its capabilities, trust (verified, signed by OpenFief's CI or by the owner, refused) and phase; a read-scoped session sees /plugins and a nav link to it; the pages offer no form or button that changes a plugin; an unavailable list shows a sentence instead of failing the page; an account's transactions name their connector instance and mark a removed one; templates stay free of inline code.