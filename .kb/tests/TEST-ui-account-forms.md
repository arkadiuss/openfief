---
title: The account forms offer an OKI account's form and an IKE's earlier payments
status: active
sourceFile: internal/views/ui/accountforms_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T20:10:27.238Z'
id: TEST-ui-account-forms
type: test
---
Against the built-in UI's handler with fake core services, the accounts page offers the four OKI forms with Not an OKI account selected and posts the chosen one with CreateAccount; an IKE's change form shows its earlier payments with the currency and posts 30 000,50 as 30000.50 with UpdateAccount; the attention item for missing earlier payments links the accounts page.