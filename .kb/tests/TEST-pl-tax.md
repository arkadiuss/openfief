---
title: Belka tax, tax wrappers and the tax FX rate date
status: active
sourceFile: internal/jurisdiction/pl/tax_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:15.749Z'
id: TEST-pl-tax
type: test
---
Table tests compute Belka tax as 19% of synthetic interest rounded half up to grosze (10.20 gives 1.94, 0.50 gives 0.10, 0.02 gives 0.00), zero for zero or negative interest, zero inside IKE and IKZE, and an error for an unknown wrapper. Another table checks that the tax FX cutoff is the day before the income event, never the event day. A false pass would be a tax rounded half even: the 0.095 tie case catches it.