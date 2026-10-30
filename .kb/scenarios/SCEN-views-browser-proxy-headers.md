---
title: The plugin trusts only what the core sets on a proxied browser request
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:57.134Z'
id: SCEN-views-browser-proxy-headers
type: scenario
---
Given a browser request on the core's browser port, reached through the owner's route or a port-forward, carrying a stolen Authorization and forged X-Forwarded-* and Forwarded headers, when the core proxies it to the view plugin that holds the browser path, then the plugin sees the core's token for that instance, the browser's Host, the scheme the core saw and no client-set forwarding header.