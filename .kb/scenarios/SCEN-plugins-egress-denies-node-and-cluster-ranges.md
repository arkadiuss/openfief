---
title: The egress proxy refuses its own node and the cluster ranges the owner names
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T00:10:35.388Z'
id: SCEN-plugins-egress-denies-node-and-cluster-ranges
type: scenario
---
Given an egress proxy told its node's public address and a global IPv6 cluster range, when a plugin sends CONNECT to an allowed host that resolves to the node's address, then the proxy answers 502 and dials nothing; an address inside the named range is refused the same way, through IPv4-mapped, NAT64 and 6to4 forms too; and a range that doesn't parse stops the proxy at start.