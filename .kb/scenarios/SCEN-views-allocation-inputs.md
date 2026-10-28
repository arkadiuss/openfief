---
title: The allocation values holdings at their stored valuation and cash at the FX rate
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:25:41.127Z'
id: SCEN-views-allocation-inputs
type: scenario
---
Given a synthetic portfolio on one date with a bond and an ETF valued, a second ETF lacking a price, and cash in PLN, USD and EUR, when the allocation is asked for by each dimension, then the valued holdings count at their stored valuation in PLN, the ETF without a valuation is listed as unvalued, and each cash balance is converted at the latest FX rate on or before the date, rounded half up to grosze like holding values (valuation.ValueRounding). A cash currency with no rate fails the call. Assumes PLN is the reporting currency.