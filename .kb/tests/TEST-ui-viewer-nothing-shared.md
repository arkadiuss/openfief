---
title: 'A viewer granted nothing: the Inbox and Sleeves empty state, and the Plugins page listing the installed plugins, over a fake core'
status: active
origin:
  kind: agent
  ref: 'author answers of 2026-11-13 recorded in the S7 task brief (S6c): invitations, Inbox and Plugins pages of a viewer'
  recorded_at: '2026-11-13T23:30:00.000Z'
sourceFile: internal/views/ui/viewernothing_test.go
verification_scope: integration
id: TEST-ui-viewer-nothing-shared
type: test
---
Serves the UI over fake Connect servers and checks a viewer the core lists no account sees 'Nothing has been shared with you yet' on the Inbox and Sleeves pages with the settings navigation on the second, and that the Plugins page lists the installed plugins with digests and signers to that viewer, to a viewer with a grant and to the owner alike.