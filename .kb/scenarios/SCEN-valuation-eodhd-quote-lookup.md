---
title: EODHD finds one listing for a ticker or ISIN, or offers the instruments a ticker names
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: SCEN-valuation-eodhd-quote-lookup
type: scenario
---
Given EODHD's search results, when a ticker names one ISIN on several venues, then the US venue's listing is quoted, else the first of D-51's exchanges; when it names several ISINs, then up to ten candidates are answered and no close is fetched; a period with no close is answered with the latest close before it, and the plugin asks EODHD only with the owner's key through its egress client. Assumes made-up listings and closes.
