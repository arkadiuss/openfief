---
title: The read API reports in the pack's currency and refuses another
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:01.134Z'
id: SCEN-views-reporting-currency-pack
type: scenario
---
Given the views over stored valuations and rates in the jurisdiction pack's currency, when they start with that currency as the reporting currency, then totals, valuations and the allocation are reported in it next to each original currency; when they start with another, such as EUR under the Polish pack, then they refuse with an unsupported reporting currency error naming both (D-180, T-12).