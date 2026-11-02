---
title: A plugin on the SDK carries the core's trace
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:26.491Z'
id: SCEN-telemetry-plugins-sdk
type: scenario
---
Given a connector served with ServeConnector or NewConnectorServer and a view calling the core with CoreHTTPClient, when the core calls it within a trace, then the plugin's handler runs in that trace and its calls to the core propagate it, and with no endpoint in its environment the plugin exports nothing.