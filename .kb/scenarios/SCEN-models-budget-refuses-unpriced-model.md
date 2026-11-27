---
title: With a budget set, a call the router can't price is refused before it is sent
status: active
origin:
  kind: agent
  ref: the owner's items for alpha.2, from the reviews of main 0bfeb62
  recorded_at: '2026-11-27T08:04:38.599Z'
id: SCEN-models-budget-refuses-unpriced-model
type: scenario
---
Given a budget of USD 1 and an empty catalog, a provider that can't list its models yet, a third party's model whose entry has no price, or a model on the owner's hardware that the catalog doesn't list, when a call is made then it is refused before any provider is called, with ErrNoPrice, an audit entry of class budget and no usage recorded, and the Ask box answers FAILED_PRECONDITION with no budget refusal header; a model on the owner's hardware whose entry has no price is still free.