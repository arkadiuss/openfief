---
title: An entry names the actor, the action, the records by ID, the time and the outcome
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T19:17:43.013Z'
id: SCEN-identity-audit-entry-content
type: scenario
---
Given a change by an owner session, when its entry is written, then it holds the actor kind and the session's ID, an action from the closed list, references of kinds the action names each by ID, the time and ok, refused or failed; an entry missing a time, with an unknown action or with a reference kind the action does not name is not written.