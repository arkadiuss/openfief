---
title: An instrument's page data comes from every viewable account
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:28.105Z'
id: SCEN-views-instrument-page
type: scenario
---
Given an ISIN held under two listings in two accounts and priced by two sources, when a read caller reads the instrument, then it gets its listings with each one's price source and latest price, its asset class and exemption classes, its prices with the primary source marked, its holdings with lots and its transactions in the accounts it may view only, and its price checks; an ISIN nobody holds or prices is not found.