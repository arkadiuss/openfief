---
title: A viewer is refused every owner call and may only sign out
status: active
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: SCEN-identity-viewer-session-reads-only
type: scenario
---
In a viewer's session each owner call of the identity service (tokens, passkeys, links, invitations, audit, grants, people) fails as forbidden, and the same session through the real core is refused by the portfolio service too; Logout works. An owner's passkey check finished with a viewer's passkey is refused.