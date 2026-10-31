---
title: The SDK's connector server serves the connector contract to the core only and reassembles imported files
status: active
sourceFile: sdk/go/openfief/serve_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:13.657Z'
id: TEST-sdk-go-serve
type: test
verification_scope: integration
---
Serves a fake connector with NewConnectorServer's handler over HTTP/2 without TLS, with a fake CoreVerifier that accepts one synthetic core token, and calls it with the generated Connect client. Calls without an Authorization header, with another bearer token, with a Basic scheme or with an empty bearer are refused as Unauthenticated and never reach the connector. With the core's token, Sync echoes the batch ID and advances the watermark, and an ImportFile stream of a header and two chunks is reassembled by ReceiveFile; a file over the limit, a stream without a header and an empty stream are refused. InstanceAudience writes openfief-plugin:<instance>.