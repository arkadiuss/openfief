---
title: The owner sees and cancels pending invitations on the Access page
status: active
origin:
  kind: agent
  ref: 'author answers of 2026-11-13 recorded in the S7 task brief (S6c): invitations, Inbox and Plugins pages of a viewer'
  recorded_at: '2026-11-13T23:30:00.000Z'
id: SCEN-ui-access-pending-invites
type: scenario
---
Given a pending invitation for Dana, the People section lists Dana with the grants in words (a target that is gone by its ID) and the expiry, never the link; Cancel invitation posts one CancelInvite with no passkey check and goes back to the People section, where the invitation is gone; the core's refusal shows on the page; a viewer, a read session and a signed-out browser are refused.