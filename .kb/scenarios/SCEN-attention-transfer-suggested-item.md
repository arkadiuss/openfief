---
title: A suggested transfer raises one decision item until the owner decides
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T06:14:55.088Z'
id: SCEN-attention-transfer-suggested-item
type: scenario
---
Given a withdrawal in acc-1 and a deposit of the same amount into acc-2 within five days that the read side suggests as a transfer, when the attention checks run, then one open transfer-suggested item is raised with the pair as its subject, grouped as a decision, tied to acc-1 and acc-2 and hidden from a caller who may not view both; when the owner confirms or rejects the pair, the next check clears it.