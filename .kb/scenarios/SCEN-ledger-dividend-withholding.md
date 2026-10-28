---
title: A dividend with withholding tax is booked gross, tax and net
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:03.661Z'
id: SCEN-ledger-dividend-withholding
type: scenario
---
Given a dividend of 100 USD on an instrument with 15 USD withheld at source, when it is booked, then cash receives 85 USD, the USD withholding tax ledger account 15 USD, and the instrument's holding is the source of 100 USD income with no change in units; the postings sum to zero in USD and nothing is converted to PLN.