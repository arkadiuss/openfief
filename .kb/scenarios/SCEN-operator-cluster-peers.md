---
title: The operator refuses a core or proxy Service it cannot isolate plugins against
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:23.397Z'
id: SCEN-operator-cluster-peers
type: scenario
---
Given core and egress proxy Services with ClusterIPs, named TCP ports and selectors of their own pods, when the operator reads them, then plugins get those addresses and ports. Given a headless Service, a named target port, no selector, or a selector that also matches plugin pods, then the operator refuses the cluster and no plugin runs.