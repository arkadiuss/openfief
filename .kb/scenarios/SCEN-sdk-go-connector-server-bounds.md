---
title: The SDK's connector server bounds messages and request times
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:19.350Z'
id: SCEN-sdk-go-connector-server-bounds
type: scenario
---
Given a plugin serving the connector contract with NewConnectorServer, when even the core sends a message over 4 MiB, then the call fails with RESOURCE_EXHAUSTED and the connector never sees it; and the server caps a body at 65 MiB and sets read header, read and idle timeouts.