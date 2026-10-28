---
title: Totals per date sum stored valuations without mixing currencies
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:18:07.177Z'
id: SCEN-views-valuation-totals
type: scenario
---
Given stored valuations of several holdings on several dates, some in USD and some in PLN, each with its PLN value, when their totals are asked for, then there is one total per date in date order, holding the exact sum of the PLN values and, separately, the sum per original currency; a value said to be in PLN but in another currency is an error. A date range that ends before it starts or spans more than 3700 days is refused.