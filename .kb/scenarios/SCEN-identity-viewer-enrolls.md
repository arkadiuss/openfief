---
title: A viewer enrolls with the link, is signed in read-only with their grants, and the link is spent
status: active
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: SCEN-identity-viewer-enrolls
type: scenario
---
Opening the link and creating a passkey makes the person, the passkey under their ID, one grant per target and a read session of the viewer, with entries enrollment_link_used, person_created, passkey_added, grant_created and a sign_in by the viewer; the grants are projected to the model. A second use of the link is refused. The viewer signs in again with the same passkey and gets a read session; Authenticate gives the viewer holder with their ID.