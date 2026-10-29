---
title: A bond lot becomes a deposit and a buy of its cost on its purchase date in the connector's account
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:54.621Z'
id: SCEN-treasurybonds-lot-mapping
type: scenario
---
Given valid lots and the account the connector instance is configured for, when they are mapped, then each becomes a deposit of quantity x 100 PLN against equity and a buy of its bonds for that cost on its purchase date, keyed by series and purchase date so re-entering supersedes, and each series' terms are delivered once under an instrument whose symbol is the series code; one invalid lot, a duplicate or conflicting terms refuse the batch.