---
title: An IKE seeded from a starting balance is taxed above the earlier payments entered
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:23.965Z'
id: SCEN-views-ike-paid-in-earlier-payments
type: scenario
---
Given an IKE seeded on 2026-01-02 from an opening lot with 30000 PLN of earlier payments entered, then a deposit and a withdrawal, when what was paid in is counted, then it is 30000 on the go-live, the deposit and the withdrawal, the lot no payment; without the earlier payments it is missing, and an account with its whole history counts its deposits and withdrawals alone.