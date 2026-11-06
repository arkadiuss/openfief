---
title: An instrument's price history keeps every source
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T01:12:21.300Z'
id: SCEN-valuation-instrument-price-history
type: scenario
---
Given stored prices of two listings of an ISIN from two sources, when the instrument's price history for a range is read, then every listing's prices from every source in the range come in listing, date and source order, and the latest price of each listing from each source is available on its own.