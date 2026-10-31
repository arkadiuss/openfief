---
title: A new ticker's ISIN is proposed
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:55:16.233Z'
id: SCEN-ingest-ticker-isin-lookup
type: scenario
---
Given tickers reported for the first time, when the core imports the delivery, then it asks the market-data connectors to look them up and keeps the first valid proposal of each; the eodhd connector proposes the one ISIN EODHD lists for the code on the market's supported exchange.