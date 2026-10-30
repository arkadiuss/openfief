---
title: A pack names its wrappers with the default first, and a marginal rate is a fraction above 0 and below 1
status: active
sourceFile: internal/jurisdiction/pack_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:12.334Z'
id: TEST-jurisdiction-pack
type: test
---
With the Polish pack: HasWrapper accepts taxable, ike, ikze and oki and nothing else, DefaultWrapper is taxable; ValidateMarginalRate accepts 0.12, 0.32 and 0.999 and refuses 0, a negative rate, 1 and 32.