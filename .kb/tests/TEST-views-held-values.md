---
title: HeldValues totals the owner's holdings gross per instrument
status: active
sourceFile: internal/views/app/held_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:22.565Z'
id: TEST-views-held-values
type: test
---
Tests of Views.HeldValues over the views fixture: the total with cash, the value of each instrument, the count of holdings without a price, and the refusals of nothing recorded, a caller who may view nothing and a failing store.