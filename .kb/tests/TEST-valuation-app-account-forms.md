---
title: An OKI account's form and an IKE's earlier payments are recorded only as their wrapper allows
status: active
sourceFile: internal/valuation/app/accountforms_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T20:10:09.795Z'
id: TEST-valuation-app-account-forms
type: test
---
InputRecorder.RecordAccountForm and RecordEarlierPayments over a fake recorder with the Polish pack: an OKI brokerage account's form and an IKE's 40000 PLN are recorded without any revaluation, while a form OKI doesn't have, a form for a taxable account, earlier payments of an IKZE, below zero or in euro are refused with ErrInvalidInput and record nothing.