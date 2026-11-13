---
title: The provider contract is served to the core only and translates both ways
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:17.636Z'
id: SCEN-models-provider-plugin-contract
type: scenario
---
Given the core's client and the SDK's model-provider server behind the core-token check, Describe, ListModels and Generate translate the contract's messages to the domain and back, a provider that is down is unavailable, one that rejects the request is rejected, a call without the core's token is refused, and the router reaches only an instance that runs as a model provider.