---
title: The demo dataset asks for FX rates from its earliest foreign lot's open date
status: active
sourceFile: internal/demo/demo_test.go
verification_scope: unit
verification_perspective: internal
origin:
  kind: agent
  recorded_at: '2026-10-31T12:16:55.067Z'
id: TEST-demo-dataset-fx-from
type: test
---
TestFXFromCoversEveryForeignLot generates the dataset of seed 20261028 for 2026-10-28, maps its IBKR opening Flex statement through the real ibkrflex mapper and checks that Dataset.FXFrom is on or before go-live, on or before every opening lot's open date, and equal to the earliest of them. It also fails if no IBKR lot opens before go-live, since the test would then no longer cover the gap that left the demo's net values missing.