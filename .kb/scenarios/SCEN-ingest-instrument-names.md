---
title: Connectors' names reach the ledger
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T06:59:05.564Z'
id: SCEN-ingest-instrument-names
type: scenario
---
Given an IBKR statement with trade descriptions and an XTB ticker with a name, when the sync is imported and the owner confirms the ticker's ISIN, then both names reach the ledger as delivered by their instances.