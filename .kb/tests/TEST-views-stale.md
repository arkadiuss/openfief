---
title: Holdings and lots behind a removed connector instance are marked, worked out from the log when read
status: active
sourceFile: internal/views/stale_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T00:04:44.073Z'
id: TEST-views-stale
type: test
---
Pure tests over a synthetic log: a removed ibkr-old bought ETF1, ibkr-main bought more and sold 12. The ETF1 holding is marked with ibkr-old; the old lot while open; the new lot only when the sale that cut it came from a removed instance; nothing while the installed instances are unknown or before any transaction; an oversell fails.