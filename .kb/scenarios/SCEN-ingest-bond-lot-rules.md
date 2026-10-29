---
title: A bond lot the owner enters is refused unless it can be valued on its purchase date
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:52.488Z'
id: SCEN-ingest-bond-lot-rules
type: scenario
---
Given a lot with a series code, a quantity, a purchase date and its series' terms, when it is validated as of today, then it is refused if it has no purchase date, was bought after today, has a negative margin, or pl.ValueLot cannot value it on its purchase date: an unknown bond type, a purchase date whose maturity misses the series' month, a quantity that is not a positive whole number, a negative rate, a negative or non-PLN fee.