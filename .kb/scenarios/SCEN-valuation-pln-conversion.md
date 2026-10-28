---
title: Conversion to PLN multiplies by the rate and rounds once to grosze
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:29:52.759Z'
id: SCEN-valuation-pln-conversion
type: scenario
---
Given 123.45 USD and a rate of 4.0123 PLN per USD, when the amount is converted with half-up rounding, then the result is 495.32 PLN (exactly 495.318435 before rounding) and the original USD amount is unchanged; given a rate for another currency, then the conversion fails with a currency mismatch; given an amount already in PLN, then it is only rounded to grosze.