---
title: After IBKR, XTB and bond imports the whole ledger log balances per currency
status: active
tags:
  - acceptance:p10-2
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:34.859Z'
id: SCEN-e2e-double-entry-whole-log
type: scenario
---
Given an IBKR Flex sync against the fixture, an XTB statement upload and a hand-entered bond lot, when the owner runs core check ledger in the core pod, then every version of every transaction of every account, current or superseded, sums to zero per currency and the whole log sums to zero, and every current transaction of the three accounts, as the read API serves its postings, sums to zero per currency. Runs on both kind and k3d.