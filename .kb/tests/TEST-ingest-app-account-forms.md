---
title: The owner's account entries take an OKI account's form at creation and an IKE's earlier payments
status: active
sourceFile: internal/ingest/app/accountforms_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T20:10:22.884Z'
id: TEST-ingest-app-account-forms
type: test
---
app.Accounts over in-memory fakes with the Polish pack: creating an OKI brokerage account stores its form, while an OKI account without a form or with one OKI lacks and a taxable account with a form are refused as invalid accounts and save nothing; an IKE takes 40000 PLN of earlier payments, a taxable account or a negative amount is refused. Updates move an account into OKI only with a form, keep it, and refuse another form; an IKE takes earlier payments on update.