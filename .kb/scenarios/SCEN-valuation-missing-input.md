---
title: A missing input is recorded and stops the valued-through date
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:40.322Z'
id: SCEN-valuation-missing-input
type: scenario
---
Given an ETF holding with no price on its first day, or no FX rate, when the account is revalued, then the revaluation succeeds, the affected holding valuations record the missing price or FX rate with no value, and the account is valued through the day before the first gap; given a gap before the date a later revaluation starts on, then the valued-through date stays before that gap until the missing input arrives.