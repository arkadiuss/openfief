---
title: A holding at market is converted to the pack's currency
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:21.606Z'
id: SCEN-valuation-market-holding-pack-currency
type: scenario
---
Given a holding priced in USD and FX rates of USD into PLN and into EUR, when it is valued under a pack whose reporting currency is PLN, then its value is converted with the rate into PLN and stored naming PLN; under a pack whose currency is EUR, with the rate into EUR and stored naming EUR; a rate into another base is never used (D-180).