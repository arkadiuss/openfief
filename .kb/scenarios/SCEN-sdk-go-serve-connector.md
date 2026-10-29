---
title: A connector serves the core until it is stopped
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T17:06:54.831Z'
id: SCEN-sdk-go-serve-connector
type: scenario
---
Given a plugin pod's environment with its port, instance name, the core's username and the mounted keys, when the plugin runs ServeConnector, then the core's token for the instance is served, any other caller is refused, and on SIGTERM the server stops taking calls and returns cleanly; without its port or instance name it does not start.