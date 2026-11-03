---
title: Each performance day is valued in the reporting currency
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T23:56:02.283Z'
id: SCEN-views-performance-value-reporting
type: scenario
---
Given stored valuations in the reporting currency and cash in several currencies, when the performance series is computed, then each day's value is the holdings at their valuations plus the cash converted to the reporting currency at the latest rate into it, plus money in transit; a valuation stored in another currency is an error (D-61, D-180).