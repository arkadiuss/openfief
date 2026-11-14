---
title: The core believes a forwarding header only from the proxies it is told to trust
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:37:59.543Z'
id: SCEN-views-client-address-trusted-proxy
type: scenario
---
Given the pod network as the trusted proxies and X-Forwarded-For as the header, when a request comes from a pod then the plugin gets the rightmost address that is not a proxy, when it comes from anywhere else it gets the connection's address whatever the header says, a value the browser sent in X-Openfief-Client-Ip is replaced or left out, and a header with no proxies is refused at start.