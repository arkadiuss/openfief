---
title: 'Viewer invitations: default lifetime, listing, cancelling and refusing a cancelled link, over in-memory stores'
status: active
origin:
  kind: agent
  ref: 'author answers of 2026-11-13 recorded in the S7 task brief (S6c): invitations, Inbox and Plugins pages of a viewer'
  recorded_at: '2026-11-13T23:30:00.000Z'
sourceFile: internal/identity/app/invites_test.go
verification_scope: integration
id: TEST-identity-invites-app
type: test
---
Runs the identity use cases over in-memory stores and the real WebAuthn ceremonies. TestInviteLifetimeDefaultsToAnHour checks the hour default, the 15 minute and 24 hour choices, the refusal of more, and the owner's own shorter link. TestPendingInvitesAndCancel checks the pending list, that cancelling needs no passkey check and writes one entry, that a cancelled and an expired link are refused alike with the refusal recorded, and that cancelling what is gone is not found. TestCancelledInviteCannotBeUsedMidCeremony checks a viewer who began before the cancellation is not created. TestInvitesAreTheOwners checks only an owner session lists or cancels.