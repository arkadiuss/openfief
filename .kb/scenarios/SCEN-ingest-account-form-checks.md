---
title: An OKI account's form is set when it is created and never changes
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:16.316Z'
id: SCEN-ingest-account-form-checks
type: scenario
---
Given the owner creating an OKI account, when no form or a form OKI lacks is given, then it is refused before anything is stored, and a taxable account with a form is refused too; given an account moved into OKI, it takes a form then; given one with a form, another form is refused; the form is stored with the account, served with it, and offered in the built-in UI's account forms while the account has none.