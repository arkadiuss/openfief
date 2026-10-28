---
title: Quantity of instrument units is exact
status: active
sourceFile: internal/kernel/quantity_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:19.409Z'
id: TEST-kernel-quantity
type: test
---
Table tests add and subtract fractional share quantities exactly (10.5 + 0.25 = 10.75, 3 - 3.000 = 0), negate them, compare across scales, accept a short position as a negative quantity, and reject non-numeric text with ErrInvalidDecimal. The expected values are parsed from text, never computed with floats, so a lossy implementation cannot pass.