---
title: The model portfolio page and the Inbox's proposal card
status: active
sourceFile: internal/views/ui/maintainer_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:23.260Z'
id: TEST-ui-maintainer
type: test
---
Tests of the built-in UI over a fake core on an httptest server: the page with the note and the owner's ask, what each session sees, the Inbox card with its figures, Apply and Dismiss reaching the core with the owner's session, and the refusals.