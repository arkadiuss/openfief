---
title: The owner invites a viewer with a name and grants; a bad name, an unknown target, a viewer's or demo session and a stale passkey check store no link
status: active
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: SCEN-identity-viewer-invite-link
type: scenario
---
InviteViewer with a valid name and targets that exist stores one enrollment code with the viewer's name and grants and an enrollment_link_issued entry. An empty or unprintable name, a sleeve that is not there, a viewer's session, no session and an owner session whose last passkey check is older than five minutes are each refused and store nothing.