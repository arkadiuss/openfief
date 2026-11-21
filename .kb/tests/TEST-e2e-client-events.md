---
title: On kind and k3d, a client event posted through the core's front door is taken and logged by the UI, and with the owner's toggle off it is answered 204 and dropped
status: active
origin:
  kind: agent
  ref: 'D-227 (docs/DESIGN.md): the owner can turn client events off in Settings; the author approved the deferral of the end-to-end test of /client-events (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
sourceFile: test/e2e/clientevents_test.go
verification_scope: end_to_end
id: TEST-e2e-client-events
type: test
---
Runs against the real stack (it cannot run without Docker and kind or k3d, so it is checked here to compile and vet only). An owner's session comes from an enrollment link and a virtual authenticator. A report posted with no session to the core's browser port is answered 204 and becomes a client event record in the UI's log while the login page loads events.js; the owner's post of Settings' form turns the events off, the login page stops loading the script, the settings page shows the switch off, and a second report is answered 204 and never reaches the log; turned on again, a third is logged. The test puts the setting back on when it ends.