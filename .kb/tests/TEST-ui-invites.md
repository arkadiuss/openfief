---
title: The Access page's pending invitations, the cancel form and the enrollment page's answer to a link that is gone, over a fake core
status: active
origin:
  kind: agent
  ref: 'author answers of 2026-11-13 recorded in the S7 task brief (S6c): invitations, Inbox and Plugins pages of a viewer'
  recorded_at: '2026-11-13T23:30:00.000Z'
sourceFile: internal/views/ui/invites_test.go
verification_scope: integration
id: TEST-ui-invites
type: test
---
Serves the UI over fake Connect servers and checks the pending invitations section (name, grants in words, expiry, no link), the CancelInvite request and its redirect with no passkey check, the core's refusal, the owner-only rule, and that the enrollment page answers a refused link with one plain sentence that repeats nothing the core said.