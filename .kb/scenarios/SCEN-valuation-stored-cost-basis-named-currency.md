---
title: A stored valuation keeps its cost basis and sale cost in its named currency
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:39.919Z'
id: SCEN-valuation-stored-cost-basis-named-currency
type: scenario
---
Given a holding priced at market with its net value, when its valuation is stored and read back, then its cost basis for tax and its sale cost come back in the currency the valuation names, never an assumed one (D-128, D-180).