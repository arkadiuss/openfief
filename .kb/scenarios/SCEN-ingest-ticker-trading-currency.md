---
title: An opening lot is booked in its listing's trading currency
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T16:48:13.617Z'
id: SCEN-ingest-ticker-trading-currency
type: scenario
---
Given an opening lot named by ticker with its open price and no cost, when the owner confirms its ticker's ISIN, then its cost is its quantity times the price in the currency of the confirmed listing, or of the exchange its market names when the listing names none, a price in pence converted to pounds, rounded half to even to the currency's minor unit; the market-data connector's proposal carries the listing's currency.