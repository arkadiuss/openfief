---
title: A load of another day records the loaded dataset's missing inputs and is still refused
status: active
sourceFile: cmd/demo-seed/load_test.go
verification_scope: integration
verification_perspective: internal
origin:
  kind: agent
  recorded_at: '2026-10-31T12:16:58.307Z'
id: TEST-demo-seed-top-up
type: test
---
TestLoadTopsUpADemoLoadedEarlier runs under task test:pg on a fresh database: the loader loads the dataset of 2026-10-28, then the test deletes every cross-check price and every FX rate before go-live, as a demo an older loader filled would lack them. A load for 2026-10-31 without --reset-demo must fail with the already-loaded error, leave the transaction versions and the marker's day unchanged, and have recorded FX rates before go-live again; the price checks, read through valuation's CrossChecks over the real store, must flag exactly one of IBKR's cross-check prices, on the loaded day; and the IBKR account must have a revaluation pending. checkOneFlaggedPriceCheck is shared with TestLoadEndToEnd, which asserts the same after a fresh load.