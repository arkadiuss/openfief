---
title: A range with no published FX table yields no rates and no error
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:54:17.769Z'
id: SCEN-valuation-fx-no-data
type: scenario
---
Given a date range that holds only a weekend and a public holiday, such as Saturday 2025-01-04 to Monday 2025-01-06 (Epiphany), when the rate provider is asked for its rates, then it returns an empty result rather than an error, and fetching the range returns no rates, so valuation falls back to the last earlier table.