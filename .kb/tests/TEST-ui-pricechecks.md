---
title: The built-in UI lists flagged price checks and counts them on the overview
status: active
sourceFile: internal/views/ui/pricechecks_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T07:02:59.009Z'
id: TEST-ui-pricechecks
type: test
---
Handler tests of the UI against the fake read API: GET /price-checks with a read session shows the flagged check of the synthetic instrument with both prices in GBP, both sources, 2.8232% and the 1% threshold, with no inline script or style, and without a session redirects to the login page; the overview's total card links '1 flagged price check(s)' to /price-checks and the navigation links the page; percent writes fractions as exact percentages and gives nothing for text that isn't a decimal.