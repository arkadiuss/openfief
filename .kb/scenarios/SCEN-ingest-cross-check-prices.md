---
title: The core stores a delivery's cross-check prices apart from its prices
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T06:24:29.659Z'
id: SCEN-ingest-cross-check-prices
type: scenario
---
Given a connector delivery whose cross_check_prices holds a synthetic mark of PLTEST000010/TST with a spoofed source, when the core syncs instance ibkr-main, then the mark is translated with the price rules, stored through the cross-check store with source ibkr-main, and the price store gets nothing; given a zero mark or a bad ISIN, then the delivery is refused as an invalid message; given no cross-check store, then the sync fails instead of dropping the marks.