---
title: An amount converts into the currency its caller names, rounded once
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T23:54:50.276Z'
id: SCEN-valuation-conversion-named-currency
type: scenario
---
Given an amount in USD and rates of USD into PLN and into EUR, when it is converted into PLN or into EUR, then it is multiplied exactly by the rate into that currency and rounded once to the minor unit with the stated mode; an amount already in that currency is only rounded, and a rate into another currency is never used (D-180).