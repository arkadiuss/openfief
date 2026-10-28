---
title: A broker balance that differs from the derived one is flagged with both values
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:43:53.024Z'
id: SCEN-ledger-assertion-mismatch
type: scenario
---
Given an account whose derived cash is 1280 USD at the end of 2026-03-05, when the broker reports 1287 USD for that date, then checking the balance assertion fails with a mismatch error that carries the reported 1287 and the derived 1280; when the broker reports 1287 USD for the next day, after a 7 USD dividend, then the check passes. The same holds for units of an instrument.