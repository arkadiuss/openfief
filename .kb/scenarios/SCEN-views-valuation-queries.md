---
title: Stored valuations are read per account and totalled for the portfolio
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:25:35.662Z'
id: SCEN-views-valuation-queries
type: scenario
---
Given stored valuations of an account's holdings on several dates, when the read API asks for the account's valuations from one date to another, then it gets each stored row dated in that range, inclusive, with its basis, source, original-currency value and PLN value, plus exact totals per date; when it asks for the portfolio's totals, then every account's valuations in the range are summed per date. An unknown account or a reversed range is refused before anything is loaded.