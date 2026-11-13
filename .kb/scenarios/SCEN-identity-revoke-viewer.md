---
title: Revoking a viewer ends everything they held and keeps their record
status: active
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: SCEN-identity-revoke-viewer
type: scenario
---
RevokePerson reports the sessions ended, passkeys disabled and grants ended; the viewer's session is then invalid, their login refused, their grants gone with a grant_revoked entry each, a passkey_removed entry per passkey and one person_revoked entry; ListPeople still lists them as revoked. A second revocation, an unknown ID and the owner are refused.