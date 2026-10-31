---
title: Belka tax, tax wrappers and the tax FX rate date
status: active
sourceFile: internal/jurisdiction/pl/tax_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:15.749Z'
id: TEST-pl-tax
type: test
verification_scope: unit
---
Table tests of BelkaTax (19%, half up to grosze, nothing in IKE, IKZE and OKI), the wrappers' keys taxable, ike, ikze and oki parsing back, an unknown key refused, BelkaRate and TaxRateCutoff's day before the event.