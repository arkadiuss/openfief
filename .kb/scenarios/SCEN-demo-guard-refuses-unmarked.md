---
title: The demo loader refuses a database with data and no demo marker
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:29.409Z'
id: SCEN-demo-guard-refuses-unmarked
type: scenario
---
Given a database whose ledger holds a sleeve that the demo loader did not write and that has no demo marker, when the loader runs, with or without --reset-demo, then it fails with ErrNotDemoDatabase before writing anything: the sleeve is still there and no demo, valuation or River schema was created.