---
title: Tickers and ticker transactions cross the connector contract unchanged, and invalid ones are refused
status: active
sourceFile: internal/ingest/connectrpc/tickers_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:24.037Z'
id: TEST-ingest-connectrpc-tickers
type: test
---
Writes a mapping with tickers and ticker entries, by broker account and by account, with delivery.FromMapping and reads it back with the core's deliveryFrom: the mapping must come back equal. Then it breaks the delivery one way at a time, an unknown symbol, an invalid country, a buy for a negative amount and both account fields set, and checks that each refuses the whole delivery. Pure translation, no network.