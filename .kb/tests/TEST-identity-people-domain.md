---
title: People, viewer sessions and the viewer holder in the identity domain
status: active
sourceFile: internal/identity/person_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: TEST-identity-people-domain
type: test
---
Table tests of NewViewer's name, ID and handle rules, a revocation clamped to creation and refused for the owner, a viewer's session reading only, and MayAct refusing a viewer by itself.