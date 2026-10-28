---
title: The demo loader Job completes next to the running core
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T18:01:02.335Z'
id: SCEN-e2e-demo-seed-completes
type: scenario
---
Given the demo's app layer applied with the demo-seed Job and the core starting together on an empty demo database, when the suite waits for the Job, then it completes within twelve minutes; a Failed condition or a pod that cannot start fails the wait at once.