---
title: The owner turns the client events off in Settings and the pages stop loading the script and the endpoint takes nothing
status: active
origin:
  kind: agent
  ref: 'D-227 (docs/DESIGN.md): the owner can turn client events off in Settings; the author approved the deferral of the end-to-end test of /client-events (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
id: SCEN-ui-client-events-owner-switch
type: scenario
---
Given an install with the client events on, when the owner turns them off in Settings, then every page, the login page included, loads no reporting script, a report posted to the endpoint is answered 204 and kept nowhere, and Settings shows the switch off with the one line of what is collected; when the owner turns them on again the script and the reports return.