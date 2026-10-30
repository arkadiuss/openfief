---
title: The read API's use cases give account terms, settings, totals with cash, allocations of a kind, yearly taxes and marks
status: active
sourceFile: internal/views/app/values_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T00:04:46.118Z'
id: TEST-views-app-values
type: test
---
Use-case tests over in-memory fakes with the Polish pack: the default and recorded wrappers and fee rule; the settings with the pack's format or the owner's, space-dot without a pack; 100 EUR of IKZE cash at 4.30 totalling 430 gross, 387 net and no liquidation until the marginal rate 0.32 gives 292.40; the net allocation counting it at 387 and the liquidation one leaving it out; OKI's yearly tax not in force, and under a synthetic act 3.30 on an average of 430; holding and lot marks from the plugin list.