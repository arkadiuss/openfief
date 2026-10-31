---
title: The built-in UI shows three values, value switches and marks, and writes amounts in the owner's number format
status: active
sourceFile: internal/views/ui/values_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T02:37:09.781Z'
id: TEST-ui-values
type: test
verification_scope: integration
---
Page tests through the UI handler over fake core services: the overview's total, account rows and value chart in the Polish default 5 000,00 format with a missing liquidation value; the allocation and value chart switches as pages and htmx fragments, all three lines included; the account page's wrapper, fee rule, net value, missing liquidation reason and no-longer-updated holding; picking comma-dot on the settings page and seeing it on the next page; the marginal rate in percent, refused out of range; the account forms posting the wrapper and fee rule in percent; and pure tests of the separators and labels.