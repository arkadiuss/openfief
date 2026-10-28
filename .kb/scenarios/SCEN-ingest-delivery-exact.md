---
title: A connector's mapping reaches the contract exactly
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T17:06:52.973Z'
id: SCEN-ingest-delivery-exact
type: scenario
---
Given a mapping with an opening lot of 1.5 units costing 160.50 EUR opened on 2025-11-03, a fee of 0.0001 EUR and a cash assertion of -1.20 EUR, when a connector plugin writes it as a delivery, then the amounts are 160.5, 0.0001 and -1.2 as decimal strings, the lot keeps its opening date, only the holding posting carries a quantity, and a transaction of no known kind is refused rather than sent.