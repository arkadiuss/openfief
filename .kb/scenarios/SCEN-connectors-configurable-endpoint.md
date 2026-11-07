---
title: A network connector calls the base URL its config names, and the grant approves that host
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T22:47:52.625Z'
id: SCEN-connectors-configurable-endpoint
type: scenario
---
Given a first-party network connector whose manifest names its base URL fields in egressFromConfig and requests the optional ca-bundle Secret, when the owner sets no base URL then the connector calls the provider's real host and asks the grant for nothing more; when the owner sets an https base URL on another host then the config schema accepts it, the install asks the grant for that host beside the fixed one, and the connector calls it; when the owner sets a plain http URL or one with a user, query or fragment then the config schema refuses it.