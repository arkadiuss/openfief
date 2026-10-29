---
title: EODHD closes are exact decimals in the listing's currency, pence as pounds
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:10.044Z'
id: SCEN-valuation-eodhd-closes-exact
type: scenario
---
Given an EODHD end-of-day response whose closes are JSON numbers such as 4.10120000000000000001 or 2.54E-1, out of date order, with null and zero closes, when the adapter maps it, then each kept close equals the exact decimal written, sorted by date, with null and non-positive days left out; given a listing quoted in GBX, then a close of 8412.5 becomes 84.125 GBP; given a body that is not the documented array, then the adapter fails as malformed.