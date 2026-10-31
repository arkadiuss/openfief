---
title: Tickers cross the contract checked
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:55:13.270Z'
id: SCEN-ingest-ticker-contract
type: scenario
---
Given a delivery with tickers and ticker transactions, when the core translates it, then they arrive as the connector wrote them, and a ticker transaction naming no ticker of the delivery or booking no valid transaction refuses the delivery.