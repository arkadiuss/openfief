---
title: A lot waits for its registry ISIN under its provisional one
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:25.965Z'
id: SCEN-treasurybonds-provisional-isin
type: scenario
---
Given the owner enters a lot of a series the ministry's data doesn't list an ISIN of, or while the data can't be read and the lot needs no published rate, when the connector maps it, then it is booked under the series' provisional ISIN, PL, the code, 00 and a check digit, keyed by series and purchase date; a lot of a listed series is booked under the registry's ISIN under the same keys.