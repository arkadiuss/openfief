---
title: The allocation values holdings at their stored valuation and cash at the FX rate
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:25:41.127Z'
id: SCEN-views-allocation-inputs
type: scenario
---
Given every account's holdings and cash on a date, stored valuations for some holdings on that date and FX rates to PLN, when the allocation is asked for, then each valued holding counts at its stored PLN value, each cash balance is converted at the latest rate dated on or before the date and rounded to grosze half-even, PLN cash counts as it is, and a holding with no valuation on the date is listed as unvalued; cash in a currency with no rate fails the request rather than being left out.