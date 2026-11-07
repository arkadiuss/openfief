---
title: A held series with an assumed fee raises one attention item
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:28.065Z'
id: SCEN-attention-bond-fee-assumed
type: scenario
---
Given stored series terms and accounts' latest valuations, when the attention sweep runs the check, then each series an account holds on the day it is valued through whose fee is assumed is one item, its subject the ISIN, tied to its instrument as a decision; it resolves once the owner confirms or overrides the fee or no account holds the series.