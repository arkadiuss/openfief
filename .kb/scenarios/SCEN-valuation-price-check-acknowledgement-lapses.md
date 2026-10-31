---
title: An acknowledgement lapses when a price changes
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T02:58:33.311Z'
id: SCEN-valuation-price-check-acknowledgement-lapses
type: scenario
---
Given a flagged check of a Sunday mark of 103 USD against Friday's close of 100 USD that the owner acknowledged, when the mark is delivered again at 104 USD or a newer price is compared, then the check is flagged and unacknowledged again; and a matched or currency-mismatch check cannot be acknowledged.