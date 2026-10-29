---
title: The plugin trusts only what the core sets on a proxied browser request
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:57.134Z'
id: SCEN-views-browser-proxy-headers
type: scenario
---
Given a browser request through kubectl port-forward or the owner's Ingress that carries its own Authorization, X-Forwarded-* and Forwarded headers, when the core proxies it to the view plugin, then the plugin sees the core's token in Authorization, the browser's Host, X-Forwarded-Proto as the core decided it (TLS, else the Ingress's http or https, else http) and X-Forwarded-Host, and none of the client's forwarding headers.