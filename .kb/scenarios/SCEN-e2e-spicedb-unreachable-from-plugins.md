---
title: No plugin, client or outsider pod can open a connection to SpiceDB; only the core's pod can
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T04:07:16.727Z'
id: SCEN-e2e-spicedb-unreachable-from-plugins
type: scenario
---
Given SpiceDB available and the XTB and NBP FX plugin instances running, when each plugin pod, the client pod and the outsider pod try to connect to SpiceDB's gRPC port by its Service's address, then no connection opens, while the same attempt from the core's pod connects. Runs on both kind and k3d.