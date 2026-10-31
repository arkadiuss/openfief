---
title: The views serve price checks filtered by status with their threshold
status: active
sourceFile: internal/views/app/pricechecks_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T07:02:54.538Z'
id: TEST-views-app-pricechecks
type: test
verification_scope: unit
---
Table test of Views.PriceChecks over a fake PriceChecks port holding a matched and a flagged check of the synthetic ETF1 (IETEST000026) from eodhd and ibkr-main: no status returns both with the 0.01 threshold, flagged and matched return one each, a Views built without WithPriceChecks returns no checks and no threshold, and a port failure is passed on wrapped. A false pass would need the filter to ignore the status, which the matched-only case rules out.