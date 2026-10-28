---
title: An egress host is refused as an IP literal or wildcard
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:09.404Z'
id: SCEN-plugins-egress-host-canonical
type: scenario
---
Given the egress host texts api.example.com, data.example.org:8443, 192.0.2.10, *.example.com and localhost, when each is parsed, then the first becomes api.example.com on port 443, the second keeps port 8443, the IP literal and the wildcard are rejected with their own errors, and the single-label name is rejected as an invalid hostname. Assumes the proxy, not the plugin, resolves names.