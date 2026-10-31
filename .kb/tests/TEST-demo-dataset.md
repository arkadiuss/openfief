---
title: The demo dataset is deterministic, synthetic and consistent; bond coupons are interest
status: active
sourceFile: internal/demo/demo_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:17.720Z'
id: TEST-demo-dataset
type: test
verification_scope: unit
---
Table tests generate the dataset for seed 20261028 as of 2026-10-28 and check that the same seed and date give an identical dataset under go-cmp while another seed changes the prices and another date the history; that every ISIN in the instruments, bonds, Flex statements and ledger batches is valid and carries the TESTD marker and no statement names an IBKR account but U0000000; that SyntheticISIN computes the check digits an independent Luhn calculation gives and rejects bad input; that the opening statement maps through the real ibkrflex mapper with no problems and each monthly statement with only the intended skipped SUMMARY rows; that every balance assertion of the IBKR, XTB and bond histories matches the balances the ledger derives; that the COI0128 coupon is 178.50 PLN from the published rate; that the CPI readings are the published GUS values; and that DecideLoad refuses an unmarked database with data, with or without reset. A false pass would be a mapper or ledger fake: the tests use the real ibkrflex mapper and ledger functions.