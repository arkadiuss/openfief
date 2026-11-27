---
title: A new read can't skip both the authorizer and the household-level list
status: active
origin:
  kind: agent
  ref: the owner's items for alpha.2, from the reviews of main 0bfeb62
  recorded_at: '2026-11-27T08:30:58.846Z'
id: SCEN-views-read-household-reads
type: scenario
---
Given a read procedure of Views that asks the authorizer about no account and is not on the list of household-level reads, the generative test fails and names it; given one on the list that is no read procedure, it fails too; ModelPortfolio, PriceChecks, Instrument, InstrumentPrices and Plugins are on the list, documented as household-level in the read API's page.