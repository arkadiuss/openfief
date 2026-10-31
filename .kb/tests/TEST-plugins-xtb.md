---
title: The xtb plugin maps an uploaded export through the SDK, by the XTB account, with no transfer comment
status: active
sourceFile: plugins/xtb/connector_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:28.887Z'
id: TEST-plugins-xtb
type: test
---
Serves the plugin's connector with the Go SDK's server over HTTP/2 and streams a synthetic export to it as the core, with the core's token: the delivery holds the deposit and withdrawal named by the broker account, the ticker transaction of the buy and its ticker, and no byte of the transfers' comments. No mode, an opening batch and a file that isn't an export are InvalidArgument; a sync delivers nothing and keeps the watermark; the manifest asks for the connector capability and statement files only.