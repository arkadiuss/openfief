---
title: 'The UI as a viewer: it reads with their session, changes nothing and its Ask box sees only their accounts'
status: active
sourceFile: internal/views/ui/viewer_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: TEST-ui-viewer
type: test
---
Through the UI over fake Connect servers: every read of a viewer's pages carries their token and none the plugin's, every POST route found in the source is refused with 403, and the Ask box, by command and by model tool, names only the granted account.