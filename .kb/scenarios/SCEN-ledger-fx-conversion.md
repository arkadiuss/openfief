---
title: An FX conversion balances each currency through FX ledger accounts
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:06.387Z'
id: SCEN-ledger-fx-conversion
type: scenario
---
Given 1000 USD sold for 3712.40 PLN with a 2 USD fee, when it is booked, then USD cash pays 1000 into the USD FX ledger account, the PLN FX ledger account pays 3712.40 into PLN cash, and the fee moves 2 USD from cash to fees; USD and PLN each sum to zero on their own, and no rate is stored.