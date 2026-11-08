---
title: The read API marks an instrument held under a provisional ISIN
status: active
origin:
  kind: agent
  recorded_at: '2026-11-08T02:00:38.600Z'
id: SCEN-views-instrument-isin-provisional
type: scenario
---
Given an account holds a treasury bond series under its provisional ISIN, PL, the series code, 00 and a check digit, next to a bond under a registry-like ISIN and a listed ETF, when a client lists the account's holdings or gets the instrument, then only the provisional one carries isin_provisional, in the holding's Instrument and in its InstrumentDetail.