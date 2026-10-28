---
title: A plugin without granted egress reaches only the core
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:07.473Z'
id: SCEN-plugins-network-policy-no-egress
type: scenario
---
Given a connector whose grant approves no egress host, like the XTB parser, when its workload is compiled, then its NetworkPolicy allows ingress only from the core on the plugin port and egress only to the core's pod port, with no rule for the egress proxy and none for DNS, and its container gets no HTTPS_PROXY; a connector with granted hosts additionally gets egress to the proxy's pod port and HTTPS_PROXY pointing at the proxy's ClusterIP.