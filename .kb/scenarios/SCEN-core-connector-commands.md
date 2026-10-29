---
title: The core runs connector syncs in a cluster and offers bind and sync commands
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:40.482Z'
id: SCEN-core-connector-commands
type: scenario
---
Given the core in a cluster with connectors on, when it starts, then the scheduler runs in the background without affecting readiness and shutdown waits for it; when the owner runs core connector bind with an instance and a stored account, then the binding is stored; when the owner runs core connector sync, then the instance syncs now and the batch's counts are printed; outside a cluster connector sync fails; the flags parse with their defaults and refuse an invalid sync policy or mode.