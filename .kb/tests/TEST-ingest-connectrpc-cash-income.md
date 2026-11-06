---
title: Cash income reaches the core unchanged through the connector contract
status: active
sourceFile: internal/ingest/connectrpc/cashincome_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T05:50:49.758Z'
id: TEST-ingest-connectrpc-cash-income
type: test
---
Writes free-funds interest with its tax and a closed CFD's loss as a connector would (delivery.FromMapping) and reads the delivery back as the core does (deliveryFrom): the kind TRANSACTION_KIND_CASH_INCOME and the BOOK_KIND_INCOME ledger account survive the round trip with every amount, so cash income from a connector plugin is booked as cash income, not refused or turned into another kind (D-152).