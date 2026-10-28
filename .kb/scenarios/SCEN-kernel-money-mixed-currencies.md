---
title: Mixed-currency money arithmetic fails
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:25.753Z'
id: SCEN-kernel-money-mixed-currencies
type: scenario
---
Given 1 PLN and 1 USD, when they are added, subtracted or compared, then each operation returns a currency mismatch error naming PLN and USD and no amount, and nothing is converted. Assumes both codes parse as ISO 4217 currencies; conversion between currencies is valuation's job with an explicit FX rate.