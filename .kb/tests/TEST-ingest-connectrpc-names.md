---
title: Instrument names travel through the connector contract
status: active
sourceFile: internal/ingest/connectrpc/names_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T06:58:56.855Z'
id: TEST-ingest-connectrpc-names
type: test
---
delivery.FromMapping writes each ISIN's first name once; deliveryFrom reads it back into app.Delivery; a name with a malformed ISIN makes the delivery invalid.