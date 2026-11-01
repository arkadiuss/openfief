---
title: A redeemed holding leaves at its value with the tax withheld
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:14.183Z'
id: SCEN-ledger-redemption-postings
type: scenario
---
Given ten bonds redeemed at maturity for 1124.30 PLN with 23.62 PLN Belka tax withheld, when the ledger books the redemption, then the holding gives up ten units at 1124.30, withholding tax takes 23.62 and cash 1100.68; without tax all of it is cash; no units or more tax than value is refused; it is stored and read back as a redemption.