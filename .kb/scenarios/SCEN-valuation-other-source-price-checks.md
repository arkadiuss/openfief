---
title: Another source's prices check the series source's
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:53.497Z'
id: SCEN-valuation-other-source-price-checks
type: scenario
---
Given an instrument priced by EODHD, its series source, and by a second connector instance on two days, one of which EODHD lacks, when the price checks are read, then both of the second instance's prices are cross-checks against EODHD's latest price on or before their dates, and no holding is valued from either.