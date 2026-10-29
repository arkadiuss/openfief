---
title: A cross-check price beyond the threshold of the day's price is flagged
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:09.780Z'
id: SCEN-valuation-price-check
type: scenario
---
Given a synthetic eodhd close of 100 USD for PLTEST000010/TSTA on 2026-03-06 and the default 1% threshold, when IBKR's mark that day is 101 USD, then the price check has a relative difference of 0.01 and is matched; when the mark is 101.0001 USD, then it is flagged with 0.010001; when the mark is in EUR, or dated another day, then no check is made, because only the same instrument, date and currency are comparable.