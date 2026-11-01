---
title: The treasury-bonds connector books each matured lot as a redemption and pays out what the account pays out
status: active
sourceFile: internal/ingest/treasurybonds/maturities_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:09.468Z'
id: TEST-treasurybonds-maturities
type: test
---
Table tests of treasurybonds.DueMaturities, BookMaturities, MaturityCPIYears and PayoutWithdrawal with synthetic lots: a lot matured by today is due, also in an account never seeded; one matured on an opening-lot account's go-live is history. An OTS lot redeems at 100.75 a bond with Belka 1.43 on 7.50; a TOS lot bought by exchange at 99.90 redeems at 115.76 with Belka on interest plus the 1.00 discount, then a withdrawal of its 1127.47 net on the same day, keyed with :payout; inside IKE nothing is withheld. Missing readings and an unknown wrapper wait as problems. A paid-out coupon is followed by a withdrawal only from the day the setting took effect. Amounts are worked by hand in comments.