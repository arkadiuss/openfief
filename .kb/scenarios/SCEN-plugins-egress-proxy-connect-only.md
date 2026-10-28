---
title: The egress proxy tunnels only listed destinations
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:59.667Z'
id: SCEN-plugins-egress-proxy-connect-only
type: scenario
---
Given plugin sample with example.com on its egress allow-list, when it sends CONNECT example.com, then the tunnel reaches the server; when it sends CONNECT other.example.com, asks to chain through an upstream proxy, or sends a plain HTTP proxy request, then the proxy refuses with 407 and dials nothing.