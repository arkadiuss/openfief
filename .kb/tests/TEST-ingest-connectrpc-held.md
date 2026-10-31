---
title: What the core says the bound accounts hold reaches a plugin intact over the connector contract
status: active
sourceFile: internal/ingest/connectrpc/held_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:11.577Z'
id: TEST-ingest-connectrpc-held
type: test
---
Starts an in-process Connect server with a fake connector that reads the sync request with delivery.Held, as a plugin does, and calls it through the core's connectrpc.Client with two held accounts, one seeded with a holding and an opening lot, and EDO0536's terms. It compares what the connector decoded with what the core sent, field by field with go-cmp: a lossy encoding of a date, quantity, cost or rate would fail it.