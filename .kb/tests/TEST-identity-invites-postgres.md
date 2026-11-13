---
title: Pending invitations and cancelling them against PostgreSQL
status: active
origin:
  kind: agent
  ref: 'author answers of 2026-11-13 recorded in the S7 task brief (S6c): invitations, Inbox and Plugins pages of a viewer'
  recorded_at: '2026-11-13T23:30:00.000Z'
sourceFile: internal/identity/postgres/invites_test.go
verification_scope: integration
id: TEST-identity-invites-postgres
type: test
---
Runs under task test:pg. Saves viewer invitations, an expired one, an owner link and a used one, and checks PendingInvites lists only the pending ones oldest first, CancelEnrollmentCode cancels only a pending invitation and atomically, and a cancelled code can no longer be used.