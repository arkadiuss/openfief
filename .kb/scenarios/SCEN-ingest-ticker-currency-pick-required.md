---
title: A London ticker's confirmation needs the owner's pick of currency
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T23:37:57.268Z'
id: SCEN-ingest-ticker-currency-pick-required
type: scenario
---
Given a ticker on London's exchange, which quotes in pence, pounds and dollars, whose confirmed listing names no currency, when the owner confirms its ISIN without picking one, or with one London doesn't quote in, then the core refuses with INVALID_ARGUMENT naming GBX, GBP and USD, records nothing, and its rows keep waiting; when the owner picks one, the rows are booked in it. A pick where the listing or a one-currency exchange settles the currency is refused too, and the pending ticker offers London's choices.