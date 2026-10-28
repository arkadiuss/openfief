---
title: Values are reported in the configured reporting currency, PLN only for now
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:06:05.852Z'
id: SCEN-views-reporting-currency
type: scenario
---
Given the read API configured with PLN as the reporting currency, when valuations, totals or the allocation are asked for, then each value carries its original currency and its value in PLN, and totals sum only reporting-currency values; given EUR configured instead, then building the read API fails with an unsupported reporting currency error naming EUR, because FX rates are quoted against PLN only and cross rates are undecided (T-12). The pure grouping and totals take the reporting currency as a parameter and assume none.