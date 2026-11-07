---
title: The owner confirms or overrides an assumed fee
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:26.913Z'
id: SCEN-pl-assumed-fee-confirmed-by-owner
type: scenario
---
Given held series' terms with fees from the owner, the table and assumed, when the series to confirm are asked for, then only the assumed ones come back in series order, each with the question naming its series, fee and letter; confirming one keeps its fee as the owner's, confirming any other is refused, and overriding takes any PLN fee of zero or more as the owner's.