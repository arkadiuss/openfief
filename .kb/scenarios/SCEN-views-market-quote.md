---
title: A quote is looked up by ticker or ISIN and answers with its source, its day and tier T2
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: SCEN-views-market-quote
type: scenario
---
Given an owner's install with a market-data connector, when a caller asks GetMarketQuote for a ticker or an ISIN, then the answer holds the listing, the latest close and its day, the closes of the period in date order, the connector instance that answered and the tier T2; a ticker that two instruments share is answered with both as candidates and no quote; a query that is no ticker, a period out of order, in the future or over 366 days is invalid argument, an unknown instrument not found, no connector failed precondition and a connector that fails unavailable. Assumes the connector answers with made-up instruments.
