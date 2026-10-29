---
title: Price checks compare same-day, same-currency prices exactly against the threshold
status: active
sourceFile: internal/valuation/crosscheck_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:03.755Z'
id: TEST-valuation-crosscheck
type: test
---
Pure table tests on synthetic instruments PLTEST000010/TSTA and IETEST000026/TSTB: CrossCheckPrice.Validate refuses a missing instrument, date or source and a close that isn't positive; CheckThreshold accepts fractions in (0, 1) and refuses 0, negatives, 1 and 5, and the default is 0.01; CheckPrice gives the signed relative difference rounded half-even to six places, matches exactly at the threshold, flags 101.0001 against 100 at 1%, flags a gap that rounds to zero but exceeds a tiny threshold, flags pence read as pounds, and refuses another currency, instrument or date or a zero price; CheckPrices pairs by instrument and date, skips another currency and days without a price, and orders by date, instrument and source. A false pass would need the flag to follow the rounded difference, which the rounds-to-zero case rules out.