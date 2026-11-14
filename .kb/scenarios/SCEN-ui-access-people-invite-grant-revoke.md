---
title: The owner invites a viewer, changes their grants and revokes them from the Access page
status: active
origin:
  kind: agent
  ref: 'S6b task brief: the Access page''s People section and the viewers end-to-end test (D-65, D-183, D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T22:00:00.000Z'
id: SCEN-ui-access-people-invite-grant-revoke
type: scenario
---
Given the owner's session, the People section lists the owner first and each viewer with status, grants in words, sessions and passkeys, marks a viewer with no grant "no access yet" and keeps a revoked viewer listed without forms. The invitation form posts the name, link lifetime and grants as one InviteViewer, asks for a passkey check first when the core wants one, and shows the one-time link with its expiry once; a lifetime or grant out of range never reaches the core. A viewer's grants form sends exactly the boxes changed as one UpdateGrants of the person. Revoke shows what it does and reaches RevokePerson only when confirmed, with no passkey check. Every route is the owner's: a viewer, read or demo session is refused 403 and nothing reaches the core.