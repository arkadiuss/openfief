---
title: Transaction.Validate rejects unbalanced and malformed transactions
status: active
sourceFile: internal/ledger/transaction_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:23:44.927Z'
id: TEST-ledger-transaction
type: test
verification_scope: unit
---
A table test edits a balanced synthetic buy (a USD trade with a PLN fee) one field at a time. Postings one cent short in USD, or in PLN while USD balances, must give an *UnbalancedError whose Sum is exactly that cent in that currency, found with errors.As, and whose message names the currency. A single posting, a missing source, external ID, account, kind or date, a lot open date on a buy, an opening lot without one or opened after go-live, USD posted to PLN cash, a quantity on cash, an amount without currency, an empty posting and an invalid ledger account all fail with ErrInvalidTransaction. A false pass would be a sum taken across currencies: the two-currency fixture balances only per currency.