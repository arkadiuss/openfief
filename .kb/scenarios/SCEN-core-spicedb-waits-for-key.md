---
title: A core that starts before SpiceDB's key exists waits for it, fail-closed, and never crash-loops
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T13:03:11.227Z'
id: SCEN-core-spicedb-waits-for-key
type: scenario
---
Given a fresh install whose core runs with --authz=spicedb and starts before the credentials Job has created the Secret with SpiceDB's preshared key, when the key's file is missing, empty or only whitespace, then the core does not exit: it answers /healthz, refuses every authorization check, the owner's included, never answering from the policy or with the owner's view, reports not ready on /readyz so no Service routes reads to it, logs once that it waits and never the key; and when the key appears in the file the core connects to SpiceDB, reconciles, runs the cutover and reports ready. Assumes the kubelet projects the optional Secret volume once the Secret exists, within its sync period.