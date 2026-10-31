---
title: Balance assertions match derived balances or report both values
status: active
sourceFile: internal/ledger/assertion_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:43:49.498Z'
id: TEST-ledger-assertion
type: test
verification_scope: unit
---
On the synthetic holding log (a 2000 USD deposit, a buy corrected to 12 units at 1200 USD, a sale, a dividend, EUR trades) CheckBalanceAssertion is table-tested: cash of 1287 USD at the end of 2026-03-06 matches, also written 1287.00; 5 EUR and 0 PLN match; 11 units of A on 2026-03-05 and 0 of the sold-out B match. 1287 USD asserted a day early, and 13 units of A asserted for the day of the corrected buy, fail with an *AssertionMismatchError found by errors.As that carries the assertion and the derived 1280 and 15, and whose message shows both. An assertion on the fee ledger account, without account, date or currency fails with ErrInvalidAssertion. CashOn is checked per date, and an invalid log fails with ErrInvalidLog.