---
title: A viewer invited by the owner reads only what was granted and loses it when revoked
status: active
origin:
  kind: agent
  ref: 'S6b task brief: the Access page''s People section and the viewers end-to-end test (D-65, D-183, D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T22:00:00.000Z'
id: SCEN-e2e-viewers-read-only-their-grants
type: scenario
---
Given the stack whose core answers from SpiceDB and an owner session, when the owner creates a script token with one sleeve grant it lists that sleeve's accounts alone; when the owner invites a viewer with the same grant, the viewer enrolls a passkey through the link (which is then spent), signs in, lists only that sleeve's accounts through the read API and the UI, is refused a change and the Access page; when the owner revokes the viewer their session stops working.