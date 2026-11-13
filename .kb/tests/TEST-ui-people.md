---
title: The People section, its invitation, grants and revoke forms, the owner-only rule and a viewer's empty state, over a fake core
status: active
origin:
  kind: agent
  ref: 'S6b task brief: the Access page''s People section and the viewers end-to-end test (D-65, D-183, D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T22:00:00.000Z'
sourceFile: internal/views/ui/people_test.go
verification_scope: integration
id: TEST-ui-people
type: test
---
Serves the UI over fake Connect servers and checks the People section's rows, marks and forms, the InviteViewer request and its one-time link, the passkey check prompt and retry, the revoke confirmation, UpdateGrants of a viewer, refusals of viewer, read and demo sessions and the empty pages of a viewer granted nothing.