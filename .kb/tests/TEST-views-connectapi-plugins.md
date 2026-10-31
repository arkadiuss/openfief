---
title: The read API lists plugins and removed-instance marks over Connect and REST as its OpenAPI document says
status: active
sourceFile: internal/views/connectapi/plugins_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:24.204Z'
id: TEST-views-connectapi-plugins
type: test
verification_scope: integration
---
Serves the read API with a fake Plugins port: GET /v1/plugins and a ListPlugins Connect call return each instance with image, digest, capabilities, trust level, signer and phase, and the response is validated against the generated OpenAPI document; a transaction page carries connectorInstance and connectorInstanceRemoved; without the port ListPlugins is UNAVAILABLE (503) with a fixed text; the SignatureTrust enum numbers match the domain's.