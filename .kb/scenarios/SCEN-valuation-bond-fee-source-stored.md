---
title: Where a series' fee came from is stored and crosses the contract
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:30.477Z'
id: SCEN-valuation-bond-fee-source-stored
type: scenario
---
Given terms whose fee is entered, from the table or assumed, when a connector delivers them and the core stores them, then the source crosses the connector contract both ways, unspecified read as entered and an unknown value refused, is stored with the terms, comes back with them and with every stored series' terms, and the owner's fee replaces only the fee and its source.