---
title: Stored valuations are served with the currency they name
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:30.730Z'
id: SCEN-views-valuation-queries-named-currency
type: scenario
---
Given stored valuations of an account, when the read API is asked for a date range, then each valuation in it is returned with its basis, source, original-currency value and the value in the reporting currency it names, with totals per date and the portfolio's totals per date (D-36, D-180).