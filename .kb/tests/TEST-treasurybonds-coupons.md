---
title: The treasury-bonds connector books each coupon due since purchase as interest on every sync, keeps CPI years, and waits on missing readings
status: active
sourceFile: internal/ingest/treasurybonds/coupons_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:07.920Z'
id: TEST-treasurybonds-coupons
type: test
---
Table tests of treasurybonds.DueCoupons, CPIYears, BookCoupons and CouponSync with synthetic lots and fake GUS and NBP sources: coupons after the account's go-live through today count, none on go-live itself, none for an account not seeded, a capitalized series or an instrument that is no bond; a bond without terms is a problem. Booked coupons are interest keyed lot:<series>:<date>:coupon:<due> with the gross the period rate gives; an unknown wrapper and a missing reading wait as problems; GUS failing fails the sync; one sync asks GUS for at most four CPI years.