---
title: A pod no policy admits cannot reach the core, nor an admitted one its plugin port
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T18:01:00.009Z'
id: SCEN-e2e-core-policy-enforced
type: scenario
---
Given the demo's stack on a cluster that enforces NetworkPolicy, when a pod no policy admits connects to the core's Service, then it gets no connection while the client pod, admitted to the browser port as the demo's Traefik is, gets /healthz; and the client pod gets no connection on the connect port, the core's plugin port, which only plugin pods reach (D-115).