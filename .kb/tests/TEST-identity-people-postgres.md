---
title: People, viewer passkeys and sessions, links and ceremonies in PostgreSQL, and the migration of what an older core stored to the owner
status: active
sourceFile: internal/identity/postgres/people_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: TEST-identity-people-postgres
type: test
---
On PostgreSQL (task test:pg): migrating through 9, storing an owner, a passkey and a session and then migrating on; the owner row on an empty install; saving a viewer, revoking one with their passkeys and sessions, the counts, the owner's last-passkey rule; the table checks a viewer's owner-scoped session, a handle-less viewer and the owner's revocation hit; viewer actors and actions in the audit log; a link's viewer and a ceremony's person round trip.