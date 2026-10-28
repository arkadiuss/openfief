---
title: Only the core's token reaches a connector
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:45.046Z'
id: SCEN-sdk-go-core-only
type: scenario
---
Given a connector served with the SDK's NewConnectorServer and a core verifier, when a call carries no Authorization bearer token, or one the verifier does not accept as the core's token for this instance, then it is refused as Unauthenticated and the connector never sees it; when it carries the core's token, the connector handles it.