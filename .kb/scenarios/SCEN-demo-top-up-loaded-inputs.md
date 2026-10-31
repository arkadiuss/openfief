---
title: Another day's load adds the inputs an older loader didn't write
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:17:05.481Z'
id: SCEN-demo-top-up-loaded-inputs
type: scenario
---
Given a demo database loaded for 2026-10-28 that lacks the cross-check prices and the FX rates before go-live, when the loader runs for 2026-10-31 without --reset-demo, then it records the valuation inputs of the 2026-10-28 dataset again, fails asking for --reset-demo, leaves the ledger and the marker unchanged, and the price checks show one flagged check while the IBKR account waits for its revaluation.