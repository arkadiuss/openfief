---
title: The owner cancels a pending invitation
status: active
origin:
  kind: agent
  ref: 'author answers of 2026-11-13 recorded in the S7 task brief (S6c): invitations, Inbox and Plugins pages of a viewer'
  recorded_at: '2026-11-13T23:30:00.000Z'
id: SCEN-identity-invite-cancel
type: scenario
---
Given three invitations of which one expired, when the owner lists them the two pending ones appear with name, grants and expiry; cancelling one needs no passkey check, writes an enrollment_link_cancelled entry and leaves the other; the cancelled link is refused with the same error an expired one gets and the refusal is recorded; an invitation cancelled after the viewer began enrolling creates nobody; a viewer, a demo session or no session can neither list nor cancel.