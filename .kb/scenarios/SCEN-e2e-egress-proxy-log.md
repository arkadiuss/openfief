---
title: The egress proxy logs each CONNECT with host, decision and instance and nothing of the request
status: active
tags:
  - acceptance:p10-16
origin:
  kind: agent
  recorded_at: '2026-11-06T20:30:00.735Z'
id: SCEN-e2e-egress-proxy-log
type: scenario
---
Given the requests the granted plugin sent through the egress proxy, among them a plain HTTP request whose path, query, a header and the body hold a canary, when the egress proxy pod log is read, then each request appears as a decision line naming the plugin instance, the proxy type, the bare host and port and the refusal; every decision line of the run holds only the instance, the host and port, the decision, its reason and timing; every CONNECT the proxy allowed during the run is to a host on its instance allow-list, so outbound traffic goes only to configured destinations; and no line holds the canary, a token or a Proxy-Authorization header (GR-12). Runs on both kind and k3d.