---
title: 'The core end to end with a viewer: invitation, enrollment, reads limited to grants, refusals and revocation'
status: active
sourceFile: cmd/core/viewers_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: TEST-core-viewers
type: test
---
On PostgreSQL with the demo dataset and the real core under the policy: the owner invites a viewer with a sleeve grant, the viewer enrolls with a virtual passkey, reads only that sleeve's account, sleeve and kind, is refused the identity and portfolio owner calls, signs in again, is listed by ListPeople and, once revoked, loses every session.