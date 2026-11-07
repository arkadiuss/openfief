---
title: Valuation inputs and the model target follow the re-key
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:31.432Z'
id: SCEN-valuation-rekey-inputs
type: scenario
---
Given a bond series' terms, exemption class, prices and model target are stored under its provisional ISIN, when it is re-keyed, then they are stored under the registry's ISIN, where the registry's own rows win and a model target's weight joins, and re-keying again changes nothing.