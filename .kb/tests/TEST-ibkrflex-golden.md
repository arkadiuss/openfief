---
title: Synthetic Flex statements map to golden transactions, assertions and problems in both ingest modes
status: active
sourceFile: internal/ingest/ibkrflex/golden_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:12.828Z'
id: TEST-ibkrflex-golden
type: test
verification_scope: unit
---
Two synthetic Flex Query statements (testdata/opening.xml and testdata/activity.xml: fake accounts U0000000 and U0000001, fake trade and transaction IDs, synthetic ISINs with valid check digits) are mapped in opening and transactions mode and rendered as stable text, compared with golden files that go test -update rewrites. The goldens pin opening lots with open dates and cost, opening cash, buys, sells, FX conversions, dividends with withholding, deposits, withdrawals, fees, holding and cash assertions on the statement's last day, and the problems of options, futures, short positions, summary rows and an unconfigured account. A second test checks every mapped transaction and assertion validates in the ledger.