---
title: A path goes to the view plugin with the longest matching browser path, never to one of two equals
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T12:02:00.357Z'
id: SCEN-views-browser-route-longest-prefix
type: scenario
---
Given running view plugins with granted browser paths, when a browser asks for a path the core doesn't serve, then the core proxies it to the plugin whose browser path is the longest prefix covering the path by whole segments; with none it answers 404, and when two plugins claim that prefix it answers 503 and serves neither.