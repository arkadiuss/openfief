---
title: The egress proxy allows only exact allow-list entries
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:12.008Z'
id: SCEN-plugins-allow-list-exact
type: scenario
---
Given an allow-list for plugin sample with api.example.com:443, when the proxy is asked for api.example.com:443, API.example.com.:443, api.example.com:80, x.api.example.com:443 and api.example.com.evil.example.net:443, then only the first two are allowed. Assumes the proxy passes the CONNECT host and port unchanged.