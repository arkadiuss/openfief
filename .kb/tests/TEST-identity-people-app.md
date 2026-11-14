---
title: Viewer invitations, enrollment, refusals, read-only sessions, people and revocation over in-memory fakes of the store and the model
status: active
sourceFile: internal/identity/app/people_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: TEST-identity-people-app
type: test
---
Runs the auth use cases over the in-memory store, the real WebAuthn ceremonies with a virtual authenticator and a fake authorization model: an invitation, a viewer's enrollment with its entries, a spent and an expired link, every owner call refused in a viewer's session, a viewer's passkey refused as the owner's check, ListPeople, UpdateGrants of a viewer and revocation with its entries and the model's removals.