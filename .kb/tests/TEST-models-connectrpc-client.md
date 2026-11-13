---
title: The core's provider client translates the contract and classifies failures
status: active
sourceFile: internal/modelrouter/connectrpc/client_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:01.277Z'
id: TEST-models-connectrpc-client
type: test
---
Calls a fake ModelProviderService over HTTP: Describe, ListModels and Generate both ways, error classes by Connect code and an instance with no address or token.