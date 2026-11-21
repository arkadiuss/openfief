---
title: The client events switch is on by default, read with no session, changed by the owner alone and audited
status: active
origin:
  kind: agent
  ref: 'D-227 (docs/DESIGN.md): the owner can turn client events off in Settings; the author approved the deferral of the end-to-end test of /client-events (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
id: SCEN-settings-client-events-switch
type: scenario
---
Given a new core, when the view plugin reads the client events setting with no session it is on; when the owner's session turns it off and on the reads follow and each change is one audit entry naming the setting; when any other session or no token tries to change it nothing changes.