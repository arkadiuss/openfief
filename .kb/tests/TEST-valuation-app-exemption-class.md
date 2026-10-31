---
title: Unclassified instruments raise attention items; the owner classifies them
status: active
sourceFile: internal/valuation/app/exemptionclass_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:09.803Z'
id: TEST-valuation-app-exemption-class
type: test
---
Revaluer.Revalue with a fake attention raiser over an OKI, a classified OKI and a taxable account checks that only the unclassified OKI holding is raised; Settings.SetExemptionClass records an offered class for a valid ISIN without revaluing and refuses another class, a bad ISIN and a non-owner.