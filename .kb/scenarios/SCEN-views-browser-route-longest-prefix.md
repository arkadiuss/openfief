---
title: A path goes to the one view plugin holding a browser path, and to none while several hold one
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T12:02:00.357Z'
id: SCEN-views-browser-route-longest-prefix
type: scenario
---
Given the view plugins that run with a granted browser path, when a browser asks for a path the core doesn't serve, then with exactly one such view the core proxies the path to it when its browser path covers the path by whole segments and answers 404 otherwise; with none it answers 404; and with more than one, which the CLI and the operator prevent (D-114), it answers 503 and proxies to none of them.