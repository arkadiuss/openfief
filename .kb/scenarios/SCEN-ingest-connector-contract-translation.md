---
title: Connector deliveries are translated at the boundary
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:54:53.533Z'
id: SCEN-ingest-connector-contract-translation
type: scenario
---
Given a delivery from a connector, when its messages are well formed, then the core's adapter translates them into ledger transactions, balance assertions, ingest problems, prices, FX rates and index readings before any use case runs; when a decimal has an exponent or grouping, a date is not YYYY-MM-DD, a currency or ISIN is malformed, a required field is missing, an enum is unspecified or the delivery is for another batch, the call fails naming the field and nothing is imported.