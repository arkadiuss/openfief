---
title: ListInvites and CancelInvite over Connect are owner-only and answer NOT_FOUND for an invitation that is gone
status: active
origin:
  kind: agent
  ref: 'author answers of 2026-11-13 recorded in the S7 task brief (S6c): invitations, Inbox and Plugins pages of a viewer'
  recorded_at: '2026-11-13T23:30:00.000Z'
sourceFile: internal/identity/connectapi/invites_test.go
verification_scope: integration
id: TEST-identity-invites-connectapi
type: test
---
Starts the AuthService on an httptest server with a generated client over an in-memory store with three invitations. TestInviteCallsAreOwnerOnly checks a demo session, no session and a caller that is no view plugin are refused and nothing is cancelled. TestListAndCancelInvites checks the pending list with name, grants and times, cancelling with a stale passkey check, the audit entry and NOT_FOUND for cancelled, expired and unknown ones.