---
title: An empty database is marked and loaded; only a marked one is reset
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:32.626Z'
id: SCEN-demo-guard-marks-and-resets
type: scenario
---
Given an empty database, when the loader runs, then it writes the demo marker and loads the dataset; given a marked demo database, when the loader runs without --reset-demo, then it loads again without reset, and with --reset-demo, then it drops the demo's ledger, valuation and River schemas and loads afresh.