---
title: The owner sees each person's access and changes a viewer's grants
status: active
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: SCEN-identity-people-listed
type: scenario
---
ListPeople returns the owner and Alice with her grants, one passkey and one session; UpdateGrants adds and removes her grants; after revocation it lists her revoked with nothing working and UpdateGrants of her is not found.