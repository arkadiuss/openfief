---
title: Sales come with their proceeds, fees and the lots they consumed FIFO
status: active
sourceFile: internal/ledger/sale_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:02.466Z'
id: TEST-ledger-sale
type: test
---
ledger.Sales over synthetic buys and sells checks that each sell is returned in date order with its proceeds, its fee and the lots it consumed oldest first with their split cost, that a withdrawal is no sale, and that selling more than held is an OversellError naming the sale.