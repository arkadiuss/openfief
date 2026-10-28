---
title: A connector instance cannot touch another instance's transactions
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:54:49.291Z'
id: SCEN-ingest-instance-scoped-keys
type: scenario
---
Given connector instance xtb-main delivers a transaction with key source acc-1, when the core imports it, then its key source becomes plugin:xtb-main:acc-1, so it can never collide with or supersede a transaction of another instance or of an in-core connector, and an instance reinstalled under the same name continues from its own transactions.