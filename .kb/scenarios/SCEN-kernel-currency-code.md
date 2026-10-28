---
title: Only ISO 4217 alphabetic codes are currencies
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:29.752Z'
id: SCEN-kernel-currency-code
type: scenario
---
Given the codes PLN, pln, PL and 985, when each is parsed as a currency, then PLN is accepted and the other three are rejected as invalid currency codes. Assumes only the format is checked, so a valid-looking code outside the current ISO list is accepted.