---
title: Registry ISINs and provisional instruments cross the connector contract
status: active
sourceFile: internal/ingest/connectrpc/registry_test.go
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:15.504Z'
id: TEST-ingest-connectrpc-registry
type: test
---
Round-trips registry ISINs from a plugin's delivery (delivery.RegistryISINs) through the core's deliveryFrom, counts them against the delivery cap and refuses a malformed ISIN or instrument as ErrInvalidDelivery; and reads the provisional instruments a sync request names back as a plugin does (delivery.Provisional), refusing a malformed one with ErrMalformedHoldings. A false pass would let a malformed registry ISIN reach the re-key.