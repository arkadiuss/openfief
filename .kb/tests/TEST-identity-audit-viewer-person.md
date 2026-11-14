---
title: An entry a viewer made names the person after the session ended and the person was revoked
status: active
sourceFile: internal/identity/app/auditview_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:46.992Z'
id: TEST-identity-audit-viewer-person
type: test
---
A viewer enrolls, the owner revokes them, their session ends, and the log still returns their entries with the person's ID for the session, which the people list, revoked ones included, resolves to the name.