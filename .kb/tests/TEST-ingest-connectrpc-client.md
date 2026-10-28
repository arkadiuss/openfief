---
title: The core's connector client calls Sync and ImportFile over Connect RPC with a fresh core token
status: active
sourceFile: internal/ingest/connectrpc/client_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:54:33.703Z'
id: TEST-ingest-connectrpc-client
type: test
---
Serves a fake connector with the generated Connect handler over HTTP/2 without TLS and calls it with connectrpc.Client and H2CClient. Sync sends the batch ID, mode and watermark with Authorization Bearer and the token the fake issuer made for the instance, and returns the translated delivery with the next watermark, keys not yet scoped. ImportFile streams the header (batch, file name, media type, mode) and the file in several ChunkSize chunks, which the connector reassembles byte for byte. Failures: an unknown instance address, a failing token issuer, a zero mode, a delivery for another batch and a connector that refuses each fail both calls.