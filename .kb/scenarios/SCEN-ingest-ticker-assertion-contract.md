---
title: Ticker holdings cross the connector contract and bad ones are refused
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:17.796Z'
id: SCEN-ingest-ticker-assertion-contract
type: scenario
---
Given a delivery whose ticker assertions name a ticker of it, an account or a broker account, a date and a non-negative quantity, when the core reads it, then each reaches the mapping unchanged, by broker account where named so; one naming no ticker of the delivery, without a date, with a negative quantity, or naming both an account and a broker account, refuses the delivery.